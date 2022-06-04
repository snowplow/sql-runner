//
// Copyright (c) 2015-2022 Snowplow Analytics Ltd. All rights reserved.
//
// This program is licensed to you under the Apache License Version 2.0,
// and you may not use this file except in compliance with the Apache License Version 2.0.
// You may obtain a copy of the Apache License Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0.
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the Apache License Version 2.0 is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the Apache License Version 2.0 for the specific language governing permissions and limitations there under.
//
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	sf "github.com/snowflakedb/gosnowflake"
)

// Specific for Snowflake db
const (
	snowplowAppName = `Snowplow_OSS`
	loginTimeout    = 5 * time.Second                // by default is 60
	multiStmtName   = "multiple statement execution" // https://github.com/snowflakedb/gosnowflake/blob/e909f00ff624a7e60d4f91718f6adc92cbd0d80f/connection.go#L57-L61
)

// SnowflakeTarget represents Snowflake as target.
type SnowflakeTarget struct {
	Target
	Client *sql.DB
	Dsn    string
}

// IsConnectable tests connection to determine whether the Snowflake target is
// connectable.
func (sft SnowflakeTarget) IsConnectable() bool {
	client := sft.Client
	err := client.Ping()
	return err == nil
}

// NewSnowflakeTarget returns a ptr to a SnowflakeTarget.
func NewSnowflakeTarget(target Target) (*SnowflakeTarget, error) {
	config := &sf.Config{
		Region:       target.Region,
		Account:      target.Account,
		User:         target.Username,
		Password:     target.Password,
		Database:     target.Database,
		Warehouse:    target.Warehouse,
		LoginTimeout: loginTimeout,
	}
	if envAppName := os.Getenv(`SNOWPLOW_SQL_RUNNER_SNOWFLAKE_APP_NAME`); envAppName != `` {
		config.Application = `Snowplow_` + envAppName
	} else {
		config.Application = snowplowAppName
	}

	configStr, err := sf.DSN(config)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("snowflake", configStr)
	if err != nil {
		return nil, err
	}

	return &SnowflakeTarget{target, db, configStr}, nil
}

// GetTarget returns the Target field of SnowflakeTarget.
func (sft SnowflakeTarget) GetTarget() Target {
	return sft.Target
}

// RunQuery runs a query against the target
func (sft SnowflakeTarget) RunQuery(query ReadyQuery, dryRun bool, showQueryOutput bool) QueryStatus {
	var affected int64 = 0
	var err error

	if dryRun {
		if sft.IsConnectable() {
			log.Printf("SUCCESS: Able to connect to target database, %s\n.", sft.Account)
		} else {
			log.Printf("ERROR: Cannot connect to target database, %s\n.", sft.Account)
		}

		return QueryStatus{query, query.Path, 0, nil}
	}

	// Enable grabbing the queryID
	queryIDChannel := make(chan string)
	ctxWithQueryIDChan := sf.WithQueryIDChan(context.Background(), queryIDChannel)

	// Kick off a goroutine to grab the queryID when we get it from the driver (there should be one queryID per script)
	var queryID string
	qID := &queryID
	go func() {
		*qID = <-queryIDChannel
	}()

	// 0 allows arbitrary number of statements
	ctx, err := sf.WithMultiStatement(ctxWithQueryIDChan, 0)
	if err != nil {
		log.Printf("ERROR: Could not initialise query script.")
		return QueryStatus{query, query.Path, 0, err}
	}
	script := query.Script

	if len(strings.TrimSpace(script)) > 0 {
		if showQueryOutput {
			rows, err := sft.Client.QueryContext(ctx, script)
			if err != nil {
				return QueryStatus{query, query.Path, int(affected), err}
			}
			defer rows.Close()

			err = printSfTable(rows)
			if err != nil {
				log.Printf("ERROR: %s.", err)
				return QueryStatus{query, query.Path, int(affected), err}
			}

			for rows.NextResultSet() {
				err = printSfTable(rows)
				if err != nil {
					log.Printf("ERROR: %s.", err)
					return QueryStatus{query, query.Path, int(affected), err}
				}
			}
		} else {
			res, err := sft.Client.ExecContext(ctx, script)
			if err != nil {
				switch err.Error() {
				// If the error message is `-00001: `, the DB failed to return accurate status. Request the status and proceed accordingly.
				case "-00001: ":
					log.Println("INFO: Encountered -1 status. Polling for query result with queryID: ", queryID)
					pollResult := pollForQueryStatus(sft, queryID)
					return QueryStatus{query, query.Path, int(affected), pollResult}
				default:
					return QueryStatus{query, query.Path, int(affected), errors.Wrap(err, fmt.Sprintf("QueryID: %s", queryID))}
				}
			}
			aff, _ := res.RowsAffected()
			affected += aff
		}
	}

	return QueryStatus{query, query.Path, int(affected), err}
}

func printSfTable(rows *sql.Rows) error {
	outputBuffer := make([][]string, 0, 10)
	cols, err := rows.Columns()
	if err != nil {
		return errors.New("Unable to read columns")
	}

	// check to prevent rows.Next() on multi-statement
	// see also: https://github.com/snowflakedb/gosnowflake/issues/365
	for _, c := range cols {
		if c == multiStmtName {
			return errors.New("Unable to showQueryOutput for multi-statement queries")
		}
	}

	vals := make([]interface{}, len(cols))
	rawResult := make([][]byte, len(cols))
	for i := range rawResult {
		vals[i] = &rawResult[i]
	}

	for rows.Next() {
		err = rows.Scan(vals...)
		if err != nil {
			return errors.New("Unable to read row")
		}

		if len(vals) > 0 {
			outputBuffer = append(outputBuffer, stringify(rawResult))
		}
	}

	if len(outputBuffer) > 0 {
		log.Printf("QUERY OUTPUT:\n")
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader(cols)
		table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
		table.SetCenterSeparator("|")

		for _, row := range outputBuffer {
			table.Append(row)
		}

		table.Render() // Send output
	}
	return nil
}

func stringify(row [][]byte) []string {
	var line []string
	for _, element := range row {
		line = append(line, fmt.Sprint(string(element)))
	}
	return line
}

// Blocking function to poll for the true status of a query which didn't return a result.
func pollForQueryStatus(sft SnowflakeTarget, queryID string) error {
	// Get the snoflake driver and open a connection
	sfd := sft.Client.Driver()
	conn, err := sfd.Open(sft.Dsn)
	if err != nil {
		return errors.Wrap(err, "Failed to open connection to poll for query result.")
	}
	// Poll Snowflake for actual query status
	for {
		qStatus, err := conn.(sf.SnowflakeConnection).GetQueryStatus(context.Background(), queryID)

		switch {
		case err != nil && strings.Contains(err.Error(), "279301:"): // The driver returns an error containing this code when the query is still running.
			break
		case err != nil:
			// Any other error is genuine, return the error.
			return err
		case qStatus != nil && qStatus.ErrorCode == "":
			// A non-nil qStatus means the query completed. If the ErrorCode field is empty string, we have no error.
			return nil
		case qStatus != nil:
			// If qStatus is non-nil but has a non-zero error code, return the relevant info as an error.
			return errors.New(qStatus.ErrorMessage)
		default:
			break
		}
		// Give it a minute before polling again.
		time.Sleep(60 * time.Second)
	}
}
