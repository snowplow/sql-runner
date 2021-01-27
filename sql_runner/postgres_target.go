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
	"crypto/tls"
	"errors"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/olekukonko/tablewriter"
	"log"
	"net"
	"os"
	"time"
)

// For Redshift queries
// Currently exportable with the intent to be used among all SQL targets
const (
	SQL_dialTimeout = 10 * time.Second
	SQL_readTimeout = 8 * time.Hour // TODO: make this user configurable
)

type PostgresTarget struct {
	Target
	Client *pg.DB
}

func (pt PostgresTarget) IsConnectable() bool {
	client := pt.Client
	var result int
	_, err := client.QueryOne(&result, "SELECT 1") // empty query to test connection
	return err == nil && result == 1
}

func NewPostgresTarget(target Target) *PostgresTarget {
	var tlsConfig *tls.Config
	if target.Ssl == true {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	db := pg.Connect(&pg.Options{
		Addr:        target.Host + ":" + target.Port,
		User:        target.Username,
		Password:    target.Password,
		Database:    target.Database,
		TLSConfig:   tlsConfig,
		DialTimeout: SQL_dialTimeout,
		ReadTimeout: SQL_readTimeout,
		Dialer: func(network, addr string) (net.Conn, error) {
			cn, err := net.DialTimeout(network, addr, SQL_dialTimeout)
			if err != nil {
				return nil, err
			}
			return cn, cn.(*net.TCPConn).SetKeepAlive(true)
		},
	})

	return &PostgresTarget{target, db}
}

func (pt PostgresTarget) GetTarget() Target {
	return pt.Target
}

// Run a query against the target
func (pt PostgresTarget) RunQuery(query ReadyQuery, dryRun bool, showQueryOutput bool) QueryStatus {
	var err error = nil
	var res orm.Result
	if dryRun {
		options := pt.Client.Options()
		address := options.Addr
		if pt.IsConnectable() {
			if VerbosityOption == MAX_VERBOSITY {
				log.Printf("SUCCESS: Able to connect to target database, %s\n.", address)
			}
		} else if VerbosityOption > 0 {
			log.Printf("ERROR: Cannot connect to target database, %s\n.", address)
		}
		return QueryStatus{query, query.Path, 0, nil}
	}

	affected := 0
	if showQueryOutput {
		var results Results
		res, err = pt.Client.Query(&results, query.Script)
		if err == nil {
			affected = res.RowsAffected()
		} else {
			if VerbosityOption > 0 {
				log.Printf("ERROR: %s.", err)
			}
			return QueryStatus{query, query.Path, int(affected), err}
		}

		err = printTable(&results)
		if err != nil {
			if VerbosityOption > 0 {
				log.Printf("ERROR: %s.", err)
			}
			return QueryStatus{query, query.Path, int(affected), err}
		}
	} else {
		res, err = pt.Client.Exec(query.Script)
		if err == nil {
			affected = res.RowsAffected()
		}
	}

	return QueryStatus{query, query.Path, affected, err}
}

func printTable(results *Results) error {
	columns := make([]string, len(results.columns))
	for k := range results.columns {
		columns[k] = results.columns[k]
	}

	if results.elements == 1 {
		if results.results[0][0] == "" {
			return nil // blank output, edge case for asserts
		}
	} else if results.elements == 0 {
		return nil // break for no output
	}

	log.Printf("QUERY OUTPUT:\n")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(columns)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")

	if len(results.columns) == 0 {
		return errors.New("Unable to read columns")
	}

	for _, row := range results.results {
		table.Append(row)
	}

	table.Render() // Send output
	return nil
}
