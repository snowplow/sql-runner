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

func (sft SnowFlakeTarget) IsConnectable() bool {
	client := sft.Client
	err := client.Ping()
	return err == nil
}

func NewSnowflakeTarget(target Target) *SnowFlakeTarget {
	var region string
	if strings.Contains(target.Region, "us-west") {
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
// One statement per API call
func (sft SnowFlakeTarget) RunQuery(query ReadyQuery, dryRun bool) QueryStatus {
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
