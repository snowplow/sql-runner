package main

import (
	bq "cloud.google.com/go/bigquery"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
	"log"
	"os"
	"strings"
)

type BigQueryTarget struct {
	Target
	Client *bq.Client
}

func (bqt BigQueryTarget) IsConnectable() bool {
	var err error = nil
	ctx := context.Background()

	client := bqt.Client
	query := client.Query("SELECT 1") // empty query to test connection

	it, err := query.Read(ctx)
	if err != nil {
		log.Printf("ERROR: Failed to perform test query: %v", err)
		return false
	}

	var row []bq.Value
	err = it.Next(&row)
	if err != nil {
		log.Printf("ERROR: Failed to read test query results: %v", err)
		return false
	}

	return fmt.Sprint(row) == "[1]"
}

func NewBigQueryTarget(target Target) *BigQueryTarget {
	projectID := target.Project
	ctx := context.Background()

	client, err := bq.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("ERROR: Failed to create client: %v", err)
	}

	return &BigQueryTarget{target, client}
}

func (bqt BigQueryTarget) GetTarget() Target {
	return bqt.Target
}

// Run a query against the target
// One statement per API call
func (bqt BigQueryTarget) RunQuery(query ReadyQuery, dryRun bool, showQueryOutput bool) QueryStatus {
	var affected int64 = 0
	var err error = nil
	var schema bq.Schema = nil
	ctx := context.Background()

	if dryRun {
		if bqt.IsConnectable() {
			log.Printf("SUCCESS: Able to connect to target database, %s.", bqt.Project)
		} else {
			log.Printf("ERROR: Cannot connect to target database, %s.", bqt.Project)
		}
		return QueryStatus{query, query.Path, 0, nil}
	}

	scripts := strings.Split(query.Script, ";")

	for _, script := range scripts {
		if len(strings.TrimSpace(script)) > 0 {
			// If showing query output, perform a dry run to get column metadata
			if showQueryOutput {
				dq := bqt.Client.Query(script)
				dq.DryRun = true
				dqJob, err := dq.Run(ctx)
				if err != nil {
					log.Printf("ERROR: Failed to dry run job: %s.", err)
					return QueryStatus{query, query.Path, int(affected), err}
				}

				schema = dqJob.LastStatus().Statistics.Details.(*bq.QueryStatistics).Schema
			}

			q := bqt.Client.Query(script)

			job, err := q.Run(ctx)
			if err != nil {
				log.Printf("ERROR: Failed to run job: %s.", err)
				return QueryStatus{query, query.Path, int(affected), err}
			}

			it, err := job.Read(ctx)
			if err != nil {
				log.Printf("ERROR: Failed to read job results: %s.", err)
				return QueryStatus{query, query.Path, int(affected), err}
			}

			if showQueryOutput {
				err = printBqTable(it, schema)
				if err != nil {
					log.Printf("ERROR: Failed to print output: %s.", err)
					return QueryStatus{query, query.Path, int(affected), err}
				}
			} else {
				queryStats := job.LastStatus().Statistics.Details.(*bq.QueryStatistics)
				aff := queryStats.NumDMLAffectedRows
				affected += aff
			}
		}
	}

	return QueryStatus{query, query.Path, int(affected), err}
}

func printBqTable(rows *bq.RowIterator, schema bq.Schema) error {
	outputBuffer := make([][]string, 0, 10)

	for {
		var row []bq.Value
		err := rows.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		outputBuffer = append(outputBuffer, bqStringify(row))
	}

	if len(outputBuffer) > 0 {
		log.Printf("QUERY OUTPUT:\n")
		table := tablewriter.NewWriter(os.Stdout)

		// Get columns from table schema
		columns := make([]string, len(schema))
		for i, field := range schema {
			columns[i] = field.Name
		}
		table.SetHeader(columns)

		table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
		table.SetCenterSeparator("|")

		for _, row := range outputBuffer {
			table.Append(row)
		}

		table.Render() // Send output
	}
	return nil
}

func bqStringify(row []bq.Value) []string {
	var line []string
	for _, element := range row {
		line = append(line, fmt.Sprint(element))
	}
	return line
}
