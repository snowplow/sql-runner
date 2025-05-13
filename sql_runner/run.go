// Copyright (c) 2015-2025 Snowplow Analytics Ltd. All rights reserved.
//
// This program is licensed to you under the Apache License Version 2.0,
// and you may not use this file except in compliance with the Apache License Version 2.0.
// You may obtain a copy of the Apache License Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0.
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the Apache License Version 2.0 is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the Apache License Version 2.0 for the specific language governing permissions and limitations there under.
package main

import (
	"bytes"
	"fmt"
	"log"
	"strings"
)

const (
	redshiftType   = "redshift"
	postgresType   = "postgres"
	postgresqlType = "postgresql"
	snowflakeType  = "snowflake"
	bigqueryType   = "bigquery"

	errorUnsupportedDbType = "Database type is unsupported"
	errorFromStepNotFound  = "The fromStep argument did not match any available steps"
	errorQueryFailedInit   = "An error occurred loading the SQL file"
	errorRunQueryNotFound  = "The runQuery argument did not match any available queries"
	errorRunQueryArgument  = "Argument for -runQuery should be in format 'step::query'"
	errorNewTargetFailure  = "Failed to create target"
)

// TargetStatus reports on any errors from running the
// playbook against a singular target.
type TargetStatus struct {
	Name   string
	Errors []error // For any errors not related to a specific step
	Steps  []StepStatus
}

// StepStatus reports on any errors from running a step.
type StepStatus struct {
	Name    string
	Index   int
	Queries []QueryStatus
}

// QueryStatus reports ony any error from a query.
type QueryStatus struct {
	Query    ReadyQuery
	Path     string
	Affected int
	Error    error
}

// ReadyStep contains a step that is ready for execution.
type ReadyStep struct {
	Name    string
	Queries []ReadyQuery
}

// ReadyQuery contains a query that is ready for execution.
type ReadyQuery struct {
	Script string
	Name   string
	Path   string
}

// Run runs a playbook of SQL scripts.
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
	if len(runQueryParts) != 2 {
		err := fmt.Errorf(errorRunQueryArgument)
		return nil, makeTargetStatuses(err, targets)
	}

	var stepName, queryName string = runQueryParts[0], runQueryParts[1]
	if stepName == "" || queryName == "" {
		err := fmt.Errorf(errorRunQueryArgument)
		return nil, makeTargetStatuses(err, targets)
	}

	steps, trimErr := trimSteps(steps, stepName, targets)
	if trimErr != nil {
		return nil, trimErr
	}

	step := steps[0] // safe
	queries := []Query{}
	for _, query := range step.Queries {
		if query.Name == queryName {
			queries = append(queries, query)
			break
		}
	}

	if len(queries) == 0 {
		err := fmt.Errorf("%s: '%s'", errorRunQueryNotFound, queryName)
		return nil, makeTargetStatuses(err, targets)
	}
	step.Queries = queries

	return []Step{step}, nil
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
		if !exists {
			err := fmt.Errorf("%s: %s", errorFromStepNotFound, fromStep)
			return nil, makeTargetStatuses(err, targets)
		}
	}
	return steps[stepIndex:], nil
}

// Helper to create the corresponding []TargetStatus given an error.
func makeTargetStatuses(err error, targets []Target) []TargetStatus {
	allStatuses := make([]TargetStatus, 0, len(targets))
	for _, tgt := range targets {
		errs := []error{err}
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
			}
			readyQueries[j] = ReadyQuery{Script: queryText, Name: query.Name, Path: queryPath}
		}
		readySteps[i] = ReadyStep{Name: step.Name, Queries: readyQueries}
	}
	return readySteps, nil
}

// Helper for a load query failed error
func loadQueryFailed(targetName string, queryPath string, err error) TargetStatus {
	errs := []error{fmt.Errorf("%s: %s: %s", errorQueryFailedInit, queryPath, err)}
	return TargetStatus{
		Name:   targetName,
		Errors: errs,
		Steps:  nil,
	}
}

// --- Running

// Route to correct database client and run
func routeAndRun(target Target, readySteps []ReadyStep, targetChan chan TargetStatus, dryRun bool, showQueryOutput bool) {
	switch strings.ToLower(target.Type) {
	case redshiftType, postgresType, postgresqlType:
		go func(tgt Target) {
			pg, err := NewPostgresTarget(tgt)
			if err != nil {
				targetChan <- newTargetFailure(tgt, err)
				return
			}
			targetChan <- runSteps(pg, readySteps, dryRun, showQueryOutput)
		}(target)
	case snowflakeType:
		go func(tgt Target) {
			snfl, err := NewSnowflakeTarget(tgt)
			if err != nil {
				targetChan <- newTargetFailure(tgt, err)
				return
			}
			targetChan <- runSteps(snfl, readySteps, dryRun, showQueryOutput)
		}(target)
	case bigqueryType:
		go func(tgt Target) {
			bq, err := NewBigQueryTarget(tgt)
			if err != nil {
				targetChan <- newTargetFailure(tgt, err)
				return
			}
			targetChan <- runSteps(bq, readySteps, dryRun, showQueryOutput)
		}(target)
	default:
		targetChan <- unsupportedDbType(target.Name, target.Type)
	}
}

// Helper for an unrecognized database type
func unsupportedDbType(targetName string, targetType string) TargetStatus {
	errs := []error{fmt.Errorf("%s: %s", errorUnsupportedDbType, targetType)}
	return TargetStatus{
		Name:   targetName,
		Errors: errs,
		Steps:  nil,
	}
}

// Helper to create TargetStatus after an error on New*Target
func newTargetFailure(target Target, err error) TargetStatus {
	errs := []error{fmt.Errorf("%s: %s: %s", errorNewTargetFailure, target.Type, err.Error())}
	return TargetStatus{
		Name:   target.Name,
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
			log.Printf("EXECUTING %s (in step %s @ %s): %s", qry.Name, stepName, dbName, qry.Path)
			queryChan <- database.RunQuery(qry, dryRun, showQueryOutput)
		}(query)
	}

	// Collect statuses from each target run
	allStatuses := make([]QueryStatus, 0)
	for i := 0; i < len(queries); i++ {
		select {
		case status := <-queryChan:
			if status.Error != nil {
				log.Printf("FAILURE: %s (step %s @ target %s), ERROR: %s\n", status.Query.Name, stepName, dbName, status.Error.Error())
			} else {
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
