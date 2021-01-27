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
	"text/template"
)

var (
	failureTemplate *template.Template
)

func init() {
	failureTemplate = template.Must(template.New("failure").Parse(`
TARGET INITIALIZATION FAILURES:{{range $status := .}}{{if $status.Errors}}
* {{$status.Name}}{{range $error := $status.Errors}}, ERRORS:
  - {{$error}}{{end}}{{end}}{{end}}
QUERY FAILURES:{{range $status := .}}{{range $step := $status.Steps}}{{range $query := $step.Queries}}{{if $query.Error}}
* Query {{$query.Query.Name}} {{$query.Path}} (in step {{$step.Name}} @ target {{$status.Name}}), ERROR:
  - {{$query.Error}}{{end}}{{end}}{{end}}{{end}}
`))
}

func review(statuses []TargetStatus) (int, string) {
	exitCode, queryCount := getExitCodeAndQueryCount(statuses)

	if exitCode == 0 {
		return exitCode, getSuccessMessage(queryCount, len(statuses))
	} else if exitCode == 8 {
		var message bytes.Buffer
		message.WriteString("WARNING: No queries to run\n")
		return exitCode, message.String()
	} else {
		return exitCode, getFailureMessage(statuses)
	}
}

// Don't use a template here as executing it could fail
func getSuccessMessage(queryCount int, targetCount int) string {
	if VerbosityOption == MAX_VERBOSITY {
		return fmt.Sprintf("SUCCESS: %d queries executed against %d targets", queryCount, targetCount)
	}
	return ""
}

// TODO: maybe would be cleaner to bubble up error from this function
func getFailureMessage(statuses []TargetStatus) string {

	var message bytes.Buffer
	if err := failureTemplate.Execute(&message, statuses); err != nil {
		if VerbosityOption > 0 {
			return fmt.Sprintf("ERROR: executing failure message template itself failed: %s", err.Error())
		} else {
			return ""
		}
	}

	return message.String()
}

// getExitCodeAndQueryCount processes statuses and returns:
// - 0 for no errors
// - 5 for target initialization errors
// - 6 for query errors
// - 7 for both types of error
// Also return the total count of query statuses we have
func getExitCodeAndQueryCount(statuses []TargetStatus) (int, int) {

	initErrors := false
	queryErrors := false
	queryCount := 0

	for _, targetStatus := range statuses {
		if targetStatus.Errors != nil {
			initErrors = true
		}
	CheckQueries:
		for _, stepStatus := range targetStatus.Steps {
			for _, queryStatus := range stepStatus.Queries {
				if queryStatus.Error != nil {
					queryErrors = true
					queryCount = 0 // Reset
					break CheckQueries
				}
				queryCount++
			}
		}
	}

	var exitCode int
	switch {
	case initErrors && queryErrors:
		exitCode = 7
	case initErrors:
		exitCode = 5
	case queryErrors:
		exitCode = 6
	case queryCount == 0:
		exitCode = 8
	default:
		exitCode = 0
	}
	return exitCode, queryCount
}
