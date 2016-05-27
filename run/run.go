//
// Copyright (c) 2015 Snowplow Analytics Ltd. All rights reserved.
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
package run

import (
	"fmt"
	"github.com/snowplow/sql-runner/playbook"
	"log"
	"strings"
)

const (
	REDSHIFT_TYPE   = "redshift"
	POSTGRES_TYPE   = "postgres"
	POSTGRESQL_TYPE = "postgresql"

	ERROR_UNSUPPORTED_DB_TYPE = "Database type is unsupported"
	ERROR_FROM_STEP_NOT_FOUND = "The fromStep argument did not match any available steps"
	ERROR_QUERY_FAILED_INIT   = "An error occured loading the SQL file"
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
func Run(pb playbook.Playbook, sp playbook.SQLProvider, fromStep string, dryRun bool) []TargetStatus {

	// Trim skippable steps from the array
	steps, trimErr := trimSteps(pb.Steps, fromStep, pb.Targets)
	if trimErr != nil {
		return trimErr
	}

	// Prepare all SQL queries
	readySteps, readyErr := loadSteps(steps, sp, pb.Variables, pb.Targets)
	if readyErr != nil {
		return readyErr
	}

	targetChan := make(chan TargetStatus, len(pb.Targets))

	// Route each target to the right db client and run
	for _, tgt := range pb.Targets {
		routeAndRun(tgt, readySteps, targetChan, dryRun)
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

// Trims skippable steps
func trimSteps(steps []playbook.Step, fromStep string, targets []playbook.Target) ([]playbook.Step, []TargetStatus) {
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
			allStatuses := make([]TargetStatus, 0)
			for _, tgt := range targets {
				status := fromStepNotFound(tgt.Name, fromStep)
				allStatuses = append(allStatuses, status)
			}
			return nil, allStatuses
		}
	}
	return steps[stepIndex:], nil
}

// Helper for a fromStep not found error
func fromStepNotFound(targetName string, fromStep string) TargetStatus {
	errs := []error{fmt.Errorf("%s: %s", ERROR_FROM_STEP_NOT_FOUND, fromStep)}
	return TargetStatus{
		Name:   targetName,
		Errors: errs,
		Steps:  nil,
	}
}

// Loads all SQL files for all Steps in the playbook ahead of time
// Fails as soon as a bad query is found
func loadSteps(steps []playbook.Step, sp playbook.SQLProvider, variables map[string]interface{}, targets []playbook.Target) ([]ReadyStep, []TargetStatus) {
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

// Route to correct database client and run
func routeAndRun(target playbook.Target, readySteps []ReadyStep, targetChan chan TargetStatus, dryRun bool) {
	switch strings.ToLower(target.Type) {
	case REDSHIFT_TYPE, POSTGRES_TYPE, POSTGRESQL_TYPE:
		go func(tgt playbook.Target) {
			pg := NewPostgresTarget(tgt)
			targetChan <- runSteps(pg, readySteps, dryRun)
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
func runSteps(database Db, steps []ReadyStep, dryRun bool) TargetStatus {

	allStatuses := make([]StepStatus, len(steps))

FailFast:
	for i, stp := range steps {
		stpIndex := i + 1
		status := runQueries(database, stpIndex, stp.Name, stp.Queries, dryRun)
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
func runQueries(database Db, stepIndex int, stepName string, queries []ReadyQuery, dryRun bool) StepStatus {

	queryChan := make(chan QueryStatus, len(queries))
	dbName := database.GetTarget().Name

	// Route each target to the right db client and run
	for _, query := range queries {
		go func(qry ReadyQuery) {
			log.Printf("EXECUTING %s (in step %s @ %s): %s", qry.Name, stepName, dbName, qry.Path)
			queryChan <- database.RunQuery(qry, dryRun)
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
