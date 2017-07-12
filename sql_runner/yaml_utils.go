//
// Copyright (c) 2015-2017 Snowplow Analytics Ltd. All rights reserved.
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
	"gopkg.in/yaml.v1"
	"regexp"
	"strings"
	"text/template"
)

var (
	// Remove the prepended :s
	rubyYamlRegex = regexp.MustCompile("^(\\s*-?\\s*):?(.*)$")
)

// Parses a playbook.yml to return the targets
// to execute against and the steps to execute
func parsePlaybookYaml(playbookBytes []byte, variables map[string]string) (Playbook, error) {
	// Define and initialize the Playbook struct
	var playbook Playbook = NewPlaybook()

	// Clean up the YAML
	cleaned := cleanYaml(playbookBytes)

	// Run the yaml through the template engine
	str, err := fillPlaybookTemplate(string(cleaned[:]), variables)

	// Unmarshal the yaml into the playbook
	err = yaml.Unmarshal([]byte(str), &playbook)

	return playbook, err
}

// Because our StorageLoader's YAML file has elements with
// : prepended (bad decision to make things easier from
// our Ruby code).
func cleanYaml(rawYaml []byte) []byte {
	var lines []string
	var buffer bytes.Buffer

	lines = strings.Split(string(rawYaml), "\n")

	for _, line := range lines {
		buffer.WriteString(rubyYamlRegex.ReplaceAllString(line, "${1}${2}\n"))
	}
	return buffer.Bytes()
}

func fillPlaybookTemplate(playbookStr string, variables map[string]string) (string, error) {
	t, err := template.New("playbook").Parse(playbookStr)
	if err != nil {
		return "", err
	}

	var filled bytes.Buffer
	if err := t.Execute(&filled, variables); err != nil {
		return "", err
	}

	return filled.String(), err
}
