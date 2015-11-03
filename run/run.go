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
	"path"
	"strings"
)

const (
	REDSHIFT_TYPE   = "redshift"
	POSTGRES_TYPE   = "postgres"
	POSTGRESQL_TYPE = "postgresql"

	ERROR_UNSUPPORTED_DB_TYPE = "Database type is unsupported"
	ERROR_FROM_STEP_NOT_FOUND = "The fromStep argument did not match any available steps"
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
	Query    playbook.Query
	Path     string
	Affected int
	Error    error
}

// Runs a playbook of SQL scripts.
//
// Handles dispatch to the appropriate
// database engine
func Run(pb playbook.Playbook, sqlroot string, fromStep string) []TargetStatus {

	// Check fromStep argument to ensure that it actually matches a step and to get the
	// index to start from
	stepIndex := 0
	if fromStep != "" {
		exists := false
		for i := 0; i < len(pb.Steps); i++ {
			if pb.Steps[i].Name == fromStep {
				exists = true
				stepIndex = i
				break;
			}
		}

		// Process failure case
		if exists == false {
			allStatuses := make([]TargetStatus, 0)
			for _, tgt := range pb.Targets {
				status := fromStepNotFound(tgt.Name, fromStep)
				allStatuses = append(allStatuses, status)
			}
			return allStatuses
		}
	}

	// Trim skippable steps from the array
	pb.Steps = pb.Steps[stepIndex:]

	targetChan := make(chan TargetStatus, len(pb.Targets))

	// Route each target to the right db client and run
	for _, tgt := range pb.Targets {
		routeAndRun(tgt, sqlroot, pb.Steps, pb.Variables, targetChan)
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

// Helper for a fromStep not found error
func fromStepNotFound(targetName string, fromStep string) TargetStatus {
	errs := []error{fmt.Errorf("%s: %s", ERROR_FROM_STEP_NOT_FOUND, fromStep)}
	return TargetStatus{
		Name:   targetName,
		Errors: errs,
		Steps:  nil,
	}
}

// Route to correct database client and run
func routeAndRun(target playbook.Target, sqlroot string, steps []playbook.Step, variables map[string]interface{}, targetChan chan TargetStatus) {
	switch strings.ToLower(target.Type) {
	case REDSHIFT_TYPE, POSTGRES_TYPE, POSTGRESQL_TYPE:
		go func(tgt playbook.Target) {
			pg := NewPostgresTarget(tgt)
			targetChan <- runSteps(pg, sqlroot, steps, variables)
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
func runSteps(database Db, sqlroot string, steps []playbook.Step, variables map[string]interface{}) TargetStatus {

	allStatuses := make([]StepStatus, len(steps))

FailFast:
	for i, stp := range steps {
		stpIndex := i + 1
		status := runQueries(database, sqlroot, stpIndex, stp.Name, stp.Queries, variables)
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
func runQueries(database Db, sqlroot string, stepIndex int, stepName string, queries []playbook.Query, variables map[string]interface{}) StepStatus {

	queryChan := make(chan QueryStatus, len(queries))
	dbName := database.GetTarget().Name

	// Route each target to the right db client and run
	for _, query := range queries {
		go func(qry playbook.Query) {
			queryPath := path.Join(sqlroot, qry.File)
			log.Printf("EXECUTING %s (in step %s @ %s): %s", qry.Name, stepName, dbName, queryPath)
			queryChan <- database.RunQuery(qry, queryPath, variables)
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
