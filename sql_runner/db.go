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
	"text/template"
)

// Db is a generalized interface to a database client.
type Db interface {
	RunQuery(ReadyQuery, bool, bool) QueryStatus
	GetTarget() Target
	IsConnectable() bool
}

// Reads the script and fills in the template
func prepareQuery(queryPath string, sp SQLProvider, template bool, variables map[string]interface{}) (string, error) {

	var script string
	var err error

	script, err = sp.GetSQL(queryPath)

	if err != nil {
		return "", err
	}

	if template {
		script, err = fillTemplate(script, variables) // Yech, mutate
		if err != nil {
			return "", err
		}
	}
	return script, nil
}

// Fills in a script which is a template
func fillTemplate(script string, variables map[string]interface{}) (string, error) {
	t, err := template.New("playbook").Funcs(TemplFuncs).Parse(script)
	if err != nil {
		return "", err
	}

	var filled bytes.Buffer
	if err := t.Execute(&filled, variables); err != nil {
		return "", err
	}
	return filled.String(), nil
}
