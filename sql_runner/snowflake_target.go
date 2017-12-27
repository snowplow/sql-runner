package main

import (
	"database/sql"
	"log"
	"time"
	sf "github.com/snowflakedb/gosnowflake"
	"strings"
)

// Specific for Snowflake db
const (
	loginTimeout = 5 * time.Second // by default is 60
)

type SnowFlakeTarget struct {
	Target
	Client *sql.DB
}

func NewSnowflakeTarget(target Target) *SnowFlakeTarget {

	configStr, err := sf.DSN(&sf.Config{
		Region:       target.Region,
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
// One statement per API call
func (sft SnowFlakeTarget) RunQuery(query ReadyQuery, dryRun bool) QueryStatus {
	var affected int64 = 0
	var err error

	if dryRun {
		return QueryStatus{query, query.Path, 0, nil}
	}

	scripts := strings.Split(query.Script, ";")

	for _, script := range scripts {
		var res sql.Result

		if len(strings.TrimSpace(script)) > 0 {

			res, err = sft.Client.Exec(script)

			if err != nil {
				return QueryStatus{query, query.Path, int(affected), err}
			} else {
				aff, _ := res.RowsAffected()
				affected += aff
			}
		}
	}

	return QueryStatus{query, query.Path, int(affected), err}
}
