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
	"math/rand"
	"os"
	"strconv"
	"text/template"
	"time"
)

var (
	// TemplFuncs is the supported template functions map.
	TemplFuncs = template.FuncMap{
		"nowWithFormat": func(format string) string {
			return time.Now().Format(format)
		},
		"systemEnv": func(env string) string {
			return os.Getenv(env)
		},
		"randomInt": func() (string, error) {
			r := rand.NewSource(time.Now().UnixNano())
			return strconv.FormatInt(r.Int63(), 10), nil
		},
		"awsChainCredentials":   awsChainCredentials,
		"awsEC2RoleCredentials": awsEC2RoleCredentials,
		"awsEnvCredentials":     awsEnvCredentials,
		"awsProfileCredentials": awsProfileCredentials,
	}
)
