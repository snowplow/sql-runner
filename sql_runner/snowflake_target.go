package main

import (
	"database/sql"
	sf "github.com/snowflakedb/gosnowflake"
	"log"
	"strings"
	"time"
	//"github.com/olekukonko/tablewriter"
	"errors"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"os"
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
// One statement per API call
func (sft SnowFlakeTarget) RunQuery(query ReadyQuery, dryRun bool, showQueryOutput bool) QueryStatus {
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
		if len(strings.TrimSpace(script)) > 0 {
			if showQueryOutput {
				rows, err := sft.Client.Query(script)
				if err != nil {
					log.Printf("ERROR: %s.", err)
					return QueryStatus{query, query.Path, int(affected), err}
				}

				err = printSfTable(rows)
				if err != nil {
					log.Printf("ERROR: %s.", err)
					return QueryStatus{query, query.Path, int(affected), err}
				}
			} else {
				res, err := sft.Client.Exec(script)

				if err != nil {
					return QueryStatus{query, query.Path, int(affected), err}
				} else {
					aff, _ := res.RowsAffected()
					affected += aff
				}
			}
		}
	}

	return QueryStatus{query, query.Path, int(affected), err}
}

func printSfTable(rows *sql.Rows) error {
	outputBuffer := make([][]string, 0, 10)
	cols, err := rows.Columns() // Remember to check err afterwards
	if err != nil {
		return errors.New("Unable to read columns")
	}

	vals := make([]interface{}, len(cols))
	for i := range cols {
		vals[i] = new(sql.RawBytes)
	}

	for rows.Next() {
		err = rows.Scan(vals...)
		if err != nil {
			return errors.New("Unable to read row")
		}

		if len(vals) > 0 {
			outputBuffer = append(outputBuffer, stringify(vals))
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

func stringify(row []interface{}) []string {
	var line []string
	for _, element := range row {
		line = append(line, fmt.Sprint(element))
	}
	return line
}
