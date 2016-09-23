//
// Copyright (c) 2015-2016 Snowplow Analytics Ltd. All rights reserved.
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
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParsePlaybookYaml(t *testing.T) {
	assert := assert.New(t)

	playbook, err := parsePlaybookYaml(nil)
	assert.Nil(err)
	assert.NotNil(playbook)
	assert.Equal(0, len(playbook.Targets))
	assert.Equal(0, len(playbook.Steps))

	playbookBytes, err1 := loadLocalFile("../integration/resources/good-postgres.yml")
	assert.Nil(err1)
	assert.NotNil(playbookBytes)

	playbook, err = parsePlaybookYaml(playbookBytes)
	assert.Nil(err)
	assert.NotNil(playbook)
	assert.Equal(2, len(playbook.Targets))
	assert.Equal(5, len(playbook.Steps))
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
