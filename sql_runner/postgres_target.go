//
// Copyright (c) 2015-2016 Snowplow Analytics Ltd. All rights reserved.
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
	"gopkg.in/pg.v5"
	"crypto/tls"
	"time"
)

// For Redshift queries
const (
	dialTimeout = 10 * time.Second
	readTimeout = 8 * time.Hour // TODO: make this user configurable
)

type PostgresTarget struct {
	Target
	Client *pg.DB
}

func NewPostgresTarget(target Target) *PostgresTarget {
	var tlsConfig *tls.Config
	if target.Ssl == true {
		// TODO: Add TLSConfig setup
	}

	db := pg.Connect(&pg.Options{
		Addr:        target.Host + ":" + target.Port,
		User:        target.Username,
		Password:    target.Password,
		Database:    target.Database,
		TLSConfig:   tlsConfig,
		DialTimeout: dialTimeout,
		ReadTimeout: readTimeout,
	})

	return &PostgresTarget{target, db}
}

func (pt PostgresTarget) GetTarget() Target {
	return pt.Target
}

// Run a query against the target
func (pt PostgresTarget) RunQuery(query ReadyQuery, dryRun bool) QueryStatus {

	if dryRun {
		return QueryStatus{query, query.Path, 0, nil}
	}

	res, err := pt.Client.Exec(query.Script)
	affected := 0
	if err == nil {
		affected = res.RowsAffected()
	}

	return QueryStatus{query, query.Path, affected, err}
}
