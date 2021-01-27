//
// Copyright (c) 2015-2021 Snowplow Analytics Ltd. All rights reserved.
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
	"errors"
	"fmt"
	"github.com/olekukonko/tablewriter"
	sf "github.com/snowflakedb/gosnowflake"
	"log"
	"os"
	"strings"
	"time"
)

// Specific for Snowflake db
const (
	loginTimeout  = 5 * time.Second                // by default is 60
	multiStmtName = "multiple statement execution" // https://github.com/snowflakedb/gosnowflake/blob/e909f00ff624a7e60d4f91718f6adc92cbd0d80f/connection.go#L57-L61
)

type SnowFlakeTarget struct {
	Target
	Client *sql.DB
}

func (sft SnowFlakeTarget) IsConnectable() bool {
	client := sft.Client
	err := client.Ping()
	return err == nil
}

func NewSnowflakeTarget(target Target) *SnowFlakeTarget {
	// Note: region connection parameter is deprecated
	var region string
	if target.Region == "us-west-1" {
		region = ""
	} else {
		region = target.Region
	}

	configStr, err := sf.DSN(&sf.Config{
		Region:       region,
		Account:      target.Account,
		User:         target.Username,
		Password:     target.Password,
		Database:     target.Database,
		Warehouse:    target.Warehouse,
		LoginTimeout: loginTimeout,
	})

	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("snowflake", configStr)

	if err != nil {
		log.Fatal(err)
	}

	return &SnowFlakeTarget{target, db}
}

func (sft SnowFlakeTarget) GetTarget() Target {
	return sft.Target
}

// Run a query against the target
func (sft SnowFlakeTarget) RunQuery(query ReadyQuery, dryRun bool, showQueryOutput bool) QueryStatus {
	var affected int64 = 0
	var err error

	if dryRun {
		if sft.IsConnectable() {
			if VerbosityOption == MAX_VERBOSITY {
				log.Printf("SUCCESS: Able to connect to target database, %s\n.", sft.Account)
			}
		} else if VerbosityOption > 0 {
			log.Printf("ERROR: Cannot connect to target database, %s\n.", sft.Account)
		}

		return QueryStatus{query, query.Path, 0, nil}
	}

	// 0 allows arbitrary number of statements
	ctx, _ := sf.WithMultiStatement(context.Background(), 0)
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
					if VerbosityOption > 0 {
						log.Printf("ERROR: %s.", err)
					}
					return QueryStatus{query, query.Path, int(affected), err}
				}
			}
		} else {
			res, err := sft.Client.ExecContext(ctx, script)
			if err != nil {
				return QueryStatus{query, query.Path, int(affected), err}
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
