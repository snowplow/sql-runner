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
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
)

func TestParsePlaybookYaml(t *testing.T) {
	assert := assert.New(t)

	playbook, err := parsePlaybookYaml(nil, nil)
	assert.Nil(err)
	assert.NotNil(playbook)
	assert.Equal(0, len(playbook.Targets))
	assert.Equal(0, len(playbook.Steps))

	playbookBytes, err1 := loadLocalFile("../integration/resources/good-postgres.yml")
	assert.Nil(err1)
	assert.NotNil(playbookBytes)

	playbook, err = parsePlaybookYaml(playbookBytes, nil)
	assert.Nil(err)
	assert.NotNil(playbook)
	assert.Equal(2, len(playbook.Targets))
	assert.Equal(6, len(playbook.Steps))
}

func TestCleanYaml(t *testing.T) {
	assert := assert.New(t)

	rawYaml := []byte(":hello: world\n:world: hello")
	cleanYamlStr := string(cleanYaml(rawYaml))
	assert.Equal("hello: world\nworld: hello\n", cleanYamlStr)

	rawYaml = []byte(":hello:\n    :world: hello")
	cleanYamlStr = string(cleanYaml(rawYaml))
	assert.Equal("hello:\n    world: hello\n", cleanYamlStr)

	cleanYamlStr = string(cleanYaml(nil))
	assert.Equal("\n", cleanYamlStr)
}

func TestTemplateYaml(t *testing.T) {
	assert := assert.New(t)

	playbookBytes, err1 := loadLocalFile("../integration/resources/good-postgres-with-template.yml")
	assert.Nil(err1)

	var m map[string]string = make(map[string]string)

	m["password"] = "qwerty123"
	m["username"] = "animoto"
	m["host"] = "theinternetz"

	playbook, err := parsePlaybookYaml(playbookBytes, CLIVariables(m))

	assert.Nil(err)

	assert.Equal("qwerty123", playbook.Targets[0].Password)
	assert.Equal("animoto", playbook.Targets[0].Username)
	assert.Equal("theinternetz", playbook.Targets[0].Host)
}

func TestParse_QueryFlag(t *testing.T) {
	testCases := []struct {
		Name     string
		Playbook string
		Expected *Playbook
	}{
		{
			Name: "simple",
			Playbook: `
:targets:
- :type:      snowflake
  :query_tag: "snowplow"
`,
			Expected: &Playbook{
				Targets: []Target{
					{
						Type:     "snowflake",
						QueryTag: "snowplow",
					},
				},
				Variables: make(map[string]interface{}),
				Steps:     nil,
			},
		},
		{
			Name: "with_escaped_quotes",
			Playbook: `
:targets:
- :type:      snowflake
  :query_tag: "{module: \"base\", steps: \"main\"}"
`,
			Expected: &Playbook{
				Targets: []Target{
					{
						Type:     "snowflake",
						QueryTag: `{module: "base", steps: "main"}`,
					},
				},
				Variables: make(map[string]interface{}),
				Steps:     nil,
			},
		},
	}

	noVars := make(map[string]string)

	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			assert := assert.New(t)

			result, err := parsePlaybookYaml([]byte(tt.Playbook), noVars)
			assert.Nil(err)
			if !reflect.DeepEqual(result, tt.Expected) {
				t.Fatalf("\nGOT:\n%s\nEXPECTED:\n%s\n",
					spew.Sdump(result),
					spew.Sdump(tt.Expected))
			}
		})
	}

}

func TestParsePlaybookYaml_compatibility(t *testing.T) {
	testCases := []struct {
		Name     string
		Playbook string
		Expected *Playbook
	}{
		{
			Name: "time",
			Playbook: `
:targets:
- :name:     test
  :database: 2022-01-01
:variables:
  :model_version: snowflake/web/1.0.1
  :start_date:    2022-01-01
`,
			Expected: &Playbook{
				Targets: []Target{
					{
						Name:     "test",
						Database: "2022-01-01",
					},
				},
				Variables: map[string]interface{}{
					"model_version": "snowflake/web/1.0.1",
					"start_date":    "2022-01-01",
				},
				Steps: nil,
			},
		},
		{
			Name: "int",
			Playbook: `
:targets:
- :name:     test
  :database: 7
:variables:
  :update_cadence_days:   7
  :lookback_window_hours: -3
  :some_float:            3.14
`,
			Expected: &Playbook{
				Targets: []Target{
					{
						Name:     "test",
						Database: string("7"),
					},
				},
				Variables: map[string]interface{}{
					"update_cadence_days":   uint64(7),
					"lookback_window_hours": int64(-3),
					"some_float":            float64(3.14),
				},
				Steps: nil,
			},
		},
		{
			Name: "bool",
			Playbook: `
:targets:
- :name:     true
  :database: test
  :ssl:      true
:variables:
  :stage_next: true
:steps:
- :name: 01-stored-procedures
  :queries:
    - :name: 01-stored-procedures
      :file: standard/00-setup/01-main/01-stored-procedures.sql
      :template: true
`,
			Expected: &Playbook{
				Targets: []Target{
					{
						Name:     string("true"),
						Database: "test",
						Ssl:      bool(true),
					},
				},
				Variables: map[string]interface{}{
					"stage_next": bool(true),
				},
				Steps: []Step{
					{
						Name: "01-stored-procedures",
						Queries: []Query{
							{
								Name:     "01-stored-procedures",
								File:     "standard/00-setup/01-main/01-stored-procedures.sql",
								Template: bool(true),
							},
						},
					},
				},
			},
		},
	}

	noVars := make(map[string]string)

	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			assert := assert.New(t)

			result, err := parsePlaybookYaml([]byte(tt.Playbook), noVars)
			assert.Nil(err)
			if !reflect.DeepEqual(result, tt.Expected) {
				t.Fatalf("\nGOT:\n%s\nEXPECTED:\n%s\n",
					spew.Sdump(result),
					spew.Sdump(tt.Expected))
			}
		})
	}
}
