//
// Copyright (c) 2015-2022 Snowplow Analytics Ltd. All rights reserved.
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

// Maps exactly onto our YAML format
type Playbook struct {
	Targets   []Target
	Variables map[string]interface{}
	Steps     []Step
}

type Target struct {
	Name, Type, Host, Database, Port, Username,
	Password, Region, Account, Warehouse, Project string
	Ssl bool
}

type Step struct {
	Name    string
	Queries []Query
}

type Query struct {
	Name, File string
	Template   bool
}

// Initialize properly the Playbook
func NewPlaybook() Playbook {
	return Playbook{Variables: make(map[string]interface{})}
}

func (p Playbook) MergeCLIVariables(variables map[string]string) Playbook {
	// TODO: Ideally this would return a new copy of the playbook to avoid
	// mutable state.
	for k, v := range variables {
		p.Variables[k] = v
	}
	return p
}
