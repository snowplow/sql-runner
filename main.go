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
package main

import (
	"fmt"
	"github.com/kardianos/osext"
	"github.com/snowplow/sql-runner/playbook"
	"github.com/snowplow/sql-runner/run"
	"log"
	"os"
	"path/filepath"
)

const (
	CLI_NAME        = "sql-runner"
	CLI_DESCRIPTION = `Run playbooks of SQL scripts in series and parallel on Redshift and Postgres`
	CLI_VERSION     = "0.3.0"

	SQLROOT_BINARY   = "BINARY"
	SQLROOT_PLAYBOOK = "PLAYBOOK"
	SQLROOT_CONSUL   = "CONSUL"
)

func main() {

	options := processFlags()

	pb, err := playbook.ParsePlaybook(options.playbook, options.consul, options.variables)
	if err != nil {
		log.Fatalf("Could not parse playbook (YAML): %s", err.Error())
	}

	statuses := run.Run(pb, options.consul, options.sqlroot, options.fromStep, options.dryRun)
	code, message := review(statuses)

	log.Printf(message)
	os.Exit(code)
}

// Parse our flags. Enforces -playbook as a
// required flag
func processFlags() Options {

	var options Options = NewOptions()
	var fs = options.GetFlagSet()
	fs.Parse(os.Args[1:])

	if options.version {
		fmt.Printf("%s version: %s\n", CLI_NAME, CLI_VERSION)
		os.Exit(0)
	}

	if len(os.Args[1:]) == 0 || options.help {
		fmt.Printf("%s version: %s\n", CLI_NAME, CLI_VERSION)
		fmt.Println(CLI_DESCRIPTION)
		fmt.Println("Usage:")
		fs.PrintDefaults()
		os.Exit(0)
	}

	if options.playbook == "" {
		fmt.Println("required flag not defined: -playbook")
		os.Exit(2)
	}

	sr, err := resolveSqlRoot(options.sqlroot, options.playbook, options.consul)
	if err != nil {
		fmt.Printf("Error resolving -sqlroot: %s\n%s\n", options.sqlroot, err)
		os.Exit(2)
	}
	options.sqlroot = sr // Yech, mutate in place

	return options
}

// Resolve the path to our SQL scripts
func resolveSqlRoot(sqlroot string, playbookPath string, consulAddress string) (string, error) {
	consulErr := fmt.Errorf("Cannot use %s option with -consul argument", sqlroot)

	switch sqlroot {
	case SQLROOT_BINARY:
		if consulAddress != "" {
			return "", consulErr
		}
		return osext.ExecutableFolder()
	case SQLROOT_PLAYBOOK:
		if consulAddress != "" {
			return "", consulErr
		}
		return filepath.Abs(filepath.Dir(playbookPath))
	case SQLROOT_CONSUL:
		return playbookPath, nil
	default:
		return sqlroot, nil
	}
}
