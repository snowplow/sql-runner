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
	"bytes"
	"fmt"
	"log"
	"strings"
)

const (
	REDSHIFT_TYPE   = "redshift"
	MYSQL_TYPE      = "mysql"
	POSTGRES_TYPE   = "postgres"
	POSTGRESQL_TYPE = "postgresql"
	SNOWFLAKE_TYPE  = "snowflake"
	BIGQUERY_TYPE   = "bigquery"

	ERROR_UNSUPPORTED_DB_TYPE = "Database type is unsupported"
	ERROR_FROM_STEP_NOT_FOUND = "The fromStep argument did not match any available steps"
	ERROR_QUERY_FAILED_INIT   = "An error occurred loading the SQL file"
	ERROR_RUN_QUERY_NOT_FOUND = "The runQuery argument did not match any available queries"
)

// Reports on any errors from running the
// playbook against a singular target
type TargetStatus struct {
	Name   string
	Errors []error // For any errors not related to a specific step
	Steps  []StepStatus
}

// Reports on any errors from running a step
type StepStatus struct {
	Name    string
	Index   int
	Queries []QueryStatus
}

// Reports ony any error from a query
type QueryStatus struct {
	Query    ReadyQuery
	Path     string
	Affected int
	Error    error
}

// Contains a step that is ready for execution
type ReadyStep struct {
	Name    string
	Queries []ReadyQuery
}

// Contains a query that is ready for execution
type ReadyQuery struct {
	Script string
	Name   string
	Path   string
}

// Runs a playbook of SQL scripts.
//
// Handles dispatch to the appropriate
// database engine
func Run(pb Playbook, sp SQLProvider, fromStep string, runQuery string, dryRun bool, fillTemplates bool, showQueryOutput bool) []TargetStatus {

	var steps []Step
	var trimErr []TargetStatus

	if runQuery != "" {
		steps, trimErr = trimToQuery(pb.Steps, runQuery, pb.Targets)
	} else {
		steps, trimErr = trimSteps(pb.Steps, fromStep, pb.Targets)
	}
	if trimErr != nil {
		return trimErr
	}

	// Prepare all SQL queries
	readySteps, readyErr := loadSteps(steps, sp, pb.Variables, pb.Targets)
	if readyErr != nil {
		return readyErr
	}

	if fillTemplates {
		for _, steps := range readySteps {
			for _, query := range steps.Queries {
				var message bytes.Buffer
				message.WriteString(fmt.Sprintf("Step name: %s\n", steps.Name))
				message.WriteString(fmt.Sprintf("Query name: %s\n", query.Name))
				message.WriteString(fmt.Sprintf("Query path: %s\n", query.Path))
				message.WriteString(query.Script)
				log.Print(message.String())
			}
		}
		allStatuses := make([]TargetStatus, 0)
		return allStatuses
	}

	targetChan := make(chan TargetStatus, len(pb.Targets))

	// Route each target to the right db client and run
	for _, tgt := range pb.Targets {
		routeAndRun(tgt, readySteps, targetChan, dryRun, showQueryOutput)
	}

	// Compose statuses from each target run
	// Duplicated in runSteps, because NOGENERICS
	allStatuses := make([]TargetStatus, 0)
	for i := 0; i < len(pb.Targets); i++ {
		select {
		case status := <-targetChan:
			allStatuses = append(allStatuses, status)
		}
	}

	return allStatuses
}

// --- Pre-run processors

// Trims down to an indivdual query
func trimToQuery(steps []Step, runQuery string, targets []Target) ([]Step, []TargetStatus) {
	runQueryParts := strings.Split(runQuery, "::")

	steps, trimErr := trimSteps(steps, runQueryParts[0], targets)
	if trimErr != nil {
		return nil, trimErr
	}
	step := steps[0]

	queries := []Query{}
	for _, query := range step.Queries {
		if query.Name == runQueryParts[1] {
			queries = append(queries, query)
			break
		}
	}

	if len(queries) == 0 {
		return nil, runQueryNotFound(targets, runQuery)
	}
	step.Queries = queries

	return []Step{step}, nil
}

// Helper for a runQuery not found error
func runQueryNotFound(targets []Target, runQuery string) []TargetStatus {
	allStatuses := make([]TargetStatus, 0)
	for _, tgt := range targets {
		errs := []error{fmt.Errorf("%s: '%s'", ERROR_RUN_QUERY_NOT_FOUND, runQuery)}
		status := TargetStatus{
			Name:   tgt.Name,
			Errors: errs,
			Steps:  nil,
		}
		allStatuses = append(allStatuses, status)
	}
	return allStatuses
}

// Trims skippable steps
func trimSteps(steps []Step, fromStep string, targets []Target) ([]Step, []TargetStatus) {
	stepIndex := 0
	if fromStep != "" {
		exists := false
		for i := 0; i < len(steps); i++ {
			if steps[i].Name == fromStep {
				exists = true
				stepIndex = i
				break
			}
		}
		if exists == false {
			return nil, fromStepNotFound(targets, fromStep)
		}
	}
	return steps[stepIndex:], nil
}

// Helper for a fromStep not found error
func fromStepNotFound(targets []Target, fromStep string) []TargetStatus {
	allStatuses := make([]TargetStatus, 0)
	for _, tgt := range targets {
		errs := []error{fmt.Errorf("%s: %s", ERROR_FROM_STEP_NOT_FOUND, fromStep)}
		status := TargetStatus{
			Name:   tgt.Name,
			Errors: errs,
			Steps:  nil,
		}
		allStatuses = append(allStatuses, status)
	}
	return allStatuses
}

// Loads all SQL files for all Steps in the playbook ahead of time
// Fails as soon as a bad query is found
func loadSteps(steps []Step, sp SQLProvider, variables map[string]interface{}, targets []Target) ([]ReadyStep, []TargetStatus) {
	sCount := len(steps)
	readySteps := make([]ReadyStep, sCount)

	for i := 0; i < sCount; i++ {
		step := steps[i]
		qCount := len(step.Queries)
		readyQueries := make([]ReadyQuery, qCount)

		for j := 0; j < qCount; j++ {
			query := step.Queries[j]
			queryText, err := prepareQuery(query.File, sp, query.Template, variables)
			queryPath := sp.ResolveKey(query.File)

			if err != nil {
				allStatuses := make([]TargetStatus, 0)
				for _, tgt := range targets {
					status := loadQueryFailed(tgt.Name, queryPath, err)
					allStatuses = append(allStatuses, status)
				}
				return nil, allStatuses
			} else {
				readyQueries[j] = ReadyQuery{Script: queryText, Name: query.Name, Path: queryPath}
			}
		}
		readySteps[i] = ReadyStep{Name: step.Name, Queries: readyQueries}
	}
	return readySteps, nil
}

// Helper for a load query failed error
func loadQueryFailed(targetName string, queryPath string, err error) TargetStatus {
	errs := []error{fmt.Errorf("%s: %s: %s", ERROR_QUERY_FAILED_INIT, queryPath, err)}
	return TargetStatus{
		Name:   targetName,
		Errors: errs,
		Steps:  nil,
	}
}

// --- Running

// Route to correct database client and run
// See https://www.golang-book.com/books/intro/10#section2 on Go channels
func routeAndRun(target Target, readySteps []ReadyStep, targetChan chan TargetStatus, dryRun bool, showQueryOutput bool) {
	switch strings.ToLower(target.Type) {
	case MYSQL_TYPE:
		go func(tgt Target) {
			mys := NewMySQLTarget(tgt)
			targetChan <- runSteps(mys, readySteps, dryRun, showQueryOutput)
		}(target)
	case REDSHIFT_TYPE, POSTGRES_TYPE, POSTGRESQL_TYPE:
		go func(tgt Target) {
			pg := NewPostgresTarget(tgt)
			targetChan <- runSteps(pg, readySteps, dryRun, showQueryOutput)
		}(target)
	case SNOWFLAKE_TYPE:
		go func(tgt Target) {
			snfl := NewSnowflakeTarget(tgt)
			targetChan <- runSteps(snfl, readySteps, dryRun, showQueryOutput)
		}(target)
	case BIGQUERY_TYPE:
		go func(tgt Target) {
			bq := NewBigQueryTarget(tgt)
			targetChan <- runSteps(bq, readySteps, dryRun, showQueryOutput)
		}(target)
	default:
		targetChan <- unsupportedDbType(target.Name, target.Type)
	}
}

// Helper for an unrecognized database type
func unsupportedDbType(targetName string, targetType string) TargetStatus {
	errs := []error{fmt.Errorf("%s: %s", ERROR_UNSUPPORTED_DB_TYPE, targetType)}
	return TargetStatus{
		Name:   targetName,
		Errors: errs,
		Steps:  nil,
	}
}

// Handles the sequential flow of steps (some of
// which may involve multiple queries in parallel).
//
// runSteps fails fast - we stop executing SQL on
// this target when a step fails.
func runSteps(database Db, steps []ReadyStep, dryRun bool, showQueryOutput bool) TargetStatus {

	allStatuses := make([]StepStatus, len(steps))

FailFast:
	for i, stp := range steps {
		stpIndex := i + 1
		status := runQueries(database, stpIndex, stp.Name, stp.Queries, dryRun, showQueryOutput)
		allStatuses = append(allStatuses, status)

		for _, qry := range status.Queries {
			if qry.Error != nil {
				break FailFast
			}
		}
	}
	return TargetStatus{
		Name:   database.GetTarget().Name,
		Errors: nil,
		Steps:  allStatuses,
	}
}

// Handles running N queries in parallel.
//
// runQueries composes failures across the queries
// for a given step: if one query fails, the others
// will still complete.
func runQueries(database Db, stepIndex int, stepName string, queries []ReadyQuery, dryRun bool, showQueryOutput bool) StepStatus {

	queryChan := make(chan QueryStatus, len(queries))
	dbName := database.GetTarget().Name

	// Route each target to the right db client and run
	for _, query := range queries {
		go func(qry ReadyQuery) {
			if VerbosityOption == MAX_VERBOSITY {
				log.Printf("EXECUTING %s (in step %s @ %s): %s", qry.Name, stepName, dbName, qry.Path)
			}
			queryChan <- database.RunQuery(qry, dryRun, showQueryOutput)
		}(query)
	}

	// Collect statuses from each target run
	allStatuses := make([]QueryStatus, 0)
	for i := 0; i < len(queries); i++ {
		select {
		case status := <-queryChan:
			if status.Error != nil {
				if VerbosityOption > 0 {
					log.Printf("FAILURE: %s (step %s @ target %s), ERROR: %s\n", status.Query.Name, stepName, dbName, status.Error.Error())
				}
			} else if VerbosityOption == MAX_VERBOSITY {
				log.Printf("SUCCESS: %s (step %s @ target %s), ROWS AFFECTED: %d\n", status.Query.Name, stepName, dbName, status.Affected)
			}
			allStatuses = append(allStatuses, status)
		}
	}

	return StepStatus{
		Name:    stepName,
		Index:   stepIndex,
		Queries: allStatuses,
	}
}
