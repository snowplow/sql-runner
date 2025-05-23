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

// YAMLFilePlaybookProvider represents YAML as playbook provider.
type YAMLFilePlaybookProvider struct {
	playbookPath string
	variables    map[string]string
}

// NewYAMLFilePlaybookProvider returns a ptr to YAMLFilePlaybookProvider.
func NewYAMLFilePlaybookProvider(playbookPath string, variables map[string]string) *YAMLFilePlaybookProvider {
	return &YAMLFilePlaybookProvider{
		playbookPath: playbookPath,
		variables:    variables,
	}
}

// GetPlaybook returns a ptr to a yaml playbook.
func (p YAMLFilePlaybookProvider) GetPlaybook() (*Playbook, error) {
	lines, err := loadLocalFile(p.playbookPath)
	if err != nil {
		return nil, err
	}

	return parsePlaybookYaml(lines, p.variables)
}
