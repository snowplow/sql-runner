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
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/olekukonko/tablewriter"
)

// For Redshift queries
const (
	dialTimeout = 10 * time.Second
	readTimeout = 8 * time.Hour // TODO: make this user configurable
)

// PostgresTarget represents a Postgres as target.
type PostgresTarget struct {
	Target
	Client *pg.DB
}

// IsConnectable tests connection to determine whether the Postgres target is
// connectable.
func (pt PostgresTarget) IsConnectable() bool {
	client := pt.Client
	err := client.Ping(context.Background())

	return err == nil
}

// NewPostgresTarget returns a ptr to a PostgresTarget.
func NewPostgresTarget(target Target) (*PostgresTarget, error) {
	var tlsConfig *tls.Config
	if target.Ssl == true {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	if target.Host == "" || target.Port == "" || target.Username == "" || target.Database == "" {
		return nil, fmt.Errorf("missing target connection parameters")
	}

	db := pg.Connect(&pg.Options{
		Addr:        fmt.Sprintf("%s:%s", target.Host, target.Port),
		User:        target.Username,
		Password:    target.Password,
		Database:    target.Database,
		TLSConfig:   tlsConfig,
		DialTimeout: dialTimeout,
		ReadTimeout: readTimeout,
		Dialer: func(ctx context.Context, network, addr string) (net.Conn, error) {
			cn, err := net.DialTimeout(network, addr, dialTimeout)
			if err != nil {
				return nil, err
			}
			return cn, cn.(*net.TCPConn).SetKeepAlive(true)
		},
	})

	return &PostgresTarget{target, db}, nil
}

// GetTarget returns the Target field of PostgresTarget.
func (pt PostgresTarget) GetTarget() Target {
	return pt.Target
}

// RunQuery runs a query against the target.
func (pt PostgresTarget) RunQuery(query ReadyQuery, dryRun bool, showQueryOutput bool) QueryStatus {
	var err error = nil
	var res orm.Result
	if dryRun {
		options := pt.Client.Options()
		address := options.Addr
		if pt.IsConnectable() {
			log.Printf("SUCCESS: Able to connect to target database, %s\n.", address)
		} else {
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
			log.Printf("ERROR: %s.", err)
			return QueryStatus{query, query.Path, int(affected), err}
		}

		err = printTable(&results)
		if err != nil {
			log.Printf("ERROR: %s.", err)
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
