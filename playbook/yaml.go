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

import (
	"bufio"
	"bytes"
	"gopkg.in/yaml.v1"
	"os"
	"regexp"
	"strings"
)

var (
	// Remove the prepended :s
	rubyYamlRegex = regexp.MustCompile("^(\\s*-?\\s*):?(.*)$")
)

// Parses a playbook.yml to return the targets
// to execute against and the steps to execute
func parsePlaybookYaml(playbookPath string, consulAddress string) (Playbook, error) {

	// Define and initialize the Playbook struct
	var playbook Playbook = NewPlaybook()
	var lines []byte
	var err error

	if consulAddress == "" {
		// Load the playbook from a local file
		lines, err = loadRubyYaml(playbookPath)
		if err != nil {
			return Playbook{}, err
		}
	} else {
		// Load the playbook from a consul value
		// - Use the playbookPath as the key
		var pbStr string

		pbStr, err = GetStringValueFromConsul(consulAddress, playbookPath)
		if err != nil {
			return Playbook{}, err
		}
		lines = cleanRubyYaml(strings.Split(pbStr, "\n"))
	}

	err = yaml.Unmarshal(lines, &playbook)
	if err != nil {
		return Playbook{}, err
	}

	return playbook, nil
}

// Load a Ruby-format YAML file into a byte slice.
// (Supports vanilla YAMLs too).
func loadRubyYaml(path string) ([]byte, error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}
	return cleanRubyYaml(lines), nil
}

// Because our StorageLoader's YAML file has elements with
// : prepended (bad decision to make things easier from
// our Ruby code).
func cleanRubyYaml(lines []string) []byte {
	var buffer bytes.Buffer
	for _, line := range lines {
		buffer.WriteString(rubyYamlRegex.ReplaceAllString(line, "${1}${2}\n"))
	}
	return buffer.Bytes()
}

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
