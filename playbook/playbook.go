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
package playbook

// Maps exactly onto our YAML format
type Playbook struct {
	Targets   []Target
	Variables map[string]interface{}
	Steps     []Step
}

type Target struct {
	Name, Type, Host, Database, Port, Username, Password string
	Ssl                                                  bool
}

type Step struct {
	Name    string
	Queries []Query
}

type Query struct {
	Name, File string
	Template   bool
}

// Initialize properly the YAML
func NewPlaybook() Playbook {
	return Playbook{Variables: make(map[string]interface{})}
}

// Dispatch to format-specific parser
func ParsePlaybook(playbookPath string, consulAddress string, variables map[string]string) (Playbook, error) {
	// TODO: Add TOML support?
	playbook, err := parsePlaybookYaml(playbookPath, consulAddress)
	if err == nil {
		playbook = MergeCLIVariables(playbook, variables)
	}
	return playbook, err
}

func MergeCLIVariables(playbook Playbook, variables map[string]string) Playbook {
	for k, v := range variables {
		playbook.Variables[k] = v
	}
	return playbook
}
