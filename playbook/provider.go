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
	"io/ioutil"
	"os"
)

// Get a playbook from the adequate source
func getPlaybook(playbookPath string, consulAddress string) ([]byte, error) {
	if consulAddress == "" {
		// Load the playbook from a local file
		return loadLocalFile(playbookPath)
	} else {
		// Load the playbook from a consul value
		// - Use the playbookPath as the key
		return GetBytesFromConsul(consulAddress, playbookPath)
	}
}

// Parses a playbook.yml to return the targets
// to execute against and the steps to execute
func getAndParsePlaybookYaml(playbookPath string, consulAddress string) (Playbook, error) {

	// Define and initialize the Playbook struct
	var playbook Playbook = NewPlaybook()

	lines, err := getPlaybook(playbookPath, consulAddress)
	if err != nil {
		return playbook, err
	}

	return parsePlaybookYaml(lines)
}

// readLines reads a whole file into memory
// and returns a slice of its lines.
func loadLocalFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return ioutil.ReadAll(file)
}
