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
package run

import (
	"bytes"
	"github.com/snowplow/sql-runner/playbook"
	"io/ioutil"
	"os"
	"text/template"
	"time"
)

var (
	templFuncs = template.FuncMap{
		"nowWithFormat": func(format string) string {
			return time.Now().Format(format)
		},
		"systemEnv": func(env string) string {
			return os.Getenv(env)
		},
		"awsChainCredentials":   awsChainCredentials,
		"awsEC2RoleCredentials": awsEC2RoleCredentials,
		"awsEnvCredentials":     awsEnvCredentials,
		"awsProfileCredentials": awsProfileCredentials,
	}
)

// Generalized interface to a database client
type Db interface {
	RunQuery(ReadyQuery, bool) QueryStatus
	GetTarget() playbook.Target
}

// Reads the script and fills in the template
func prepareQuery(queryPath string, consulAddress string, template bool, variables map[string]interface{}) (string, error) {

	var script string
	var err error

	if consulAddress == "" {
		script, err = readScript(queryPath)
	} else {
		script, err = playbook.GetStringValueFromConsul(consulAddress, queryPath)
	}

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

// Reads the file ready for executing
func readScript(file string) (string, error) {

	scriptBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	return string(scriptBytes), nil
}

// Fills in a script which is a template
func fillTemplate(script string, variables map[string]interface{}) (string, error) {

	t, err := template.New("playbook").Funcs(templFuncs).Parse(script)
	if err != nil {
		return "", err
	}

	var filled bytes.Buffer
	if err := t.Execute(&filled, variables); err != nil {
		return "", err
	}
	return filled.String(), nil
}
