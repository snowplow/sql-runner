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
package main

import (
	"fmt"
)

// Playbook maps exactly onto our YAML format
type Playbook struct {
	Targets   []Target
	Variables map[string]interface{}
	Steps     []Step
}

// Target represents the playbook target.
type Target struct {
	Name, Type, Host, Database, Port, Username,
	Password, Region, Account, Warehouse, Project string
	QueryTag             string `yaml:"query_tag"`
	Ssl                  bool
	PrivateKeyPath       string `yaml:"private_key_path"`
	PrivateKeyPassphrase string `yaml:"private_key_passphrase"`
}

// Step represents a playbook step.
type Step struct {
	Name    string
	Queries []Query
}

// Query represents a playbook query.
type Query struct {
	Name, File string
	Template   bool
}

// NewPlaybook initializes properly the Playbook.
func NewPlaybook() Playbook {
	return Playbook{Variables: make(map[string]interface{})}
}

// MergeCLIVariables merges CLIVariables to playbook variables.
func (p Playbook) MergeCLIVariables(variables map[string]string) Playbook {
	// TODO: Ideally this would return a new copy of the playbook to avoid
	// mutable state.
	for k, v := range variables {
		p.Variables[k] = v
	}
	return p
}

// Validate provides a way to fail fast if playbook is invalid.
func (p Playbook) Validate() error {
	if p.Targets == nil || len(p.Targets) == 0 {
		return fmt.Errorf("no targets")
	}

	if p.Steps == nil || len(p.Steps) == 0 {
		return fmt.Errorf("no steps")
	}

	return nil
}
