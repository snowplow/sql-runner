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
	"errors"
	"fmt"
	"github.com/kardianos/osext"
	"github.com/snowplow/sql-runner/playbook"
	"github.com/snowplow/sql-runner/run"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	CLI_NAME        = "sql-runner"
	CLI_DESCRIPTION = `Run playbooks of SQL scripts in series and parallel on Redshift and Postgres`
	CLI_VERSION     = "0.4.0"

	SQLROOT_BINARY         = "BINARY"
	SQLROOT_PLAYBOOK       = "PLAYBOOK"
	SQLROOT_PLAYBOOK_CHILD = "PLAYBOOK_CHILD"
)

func main() {

	options := processFlags()

	// pb, err := playbook.ParsePlaybook(options.playbook, options.consul, options.variables)
	pbp, pbpErr := PlaybookProviderFromOptions(options)
	if pbpErr != nil {
		log.Fatalf("Could not determine playbook source: %s", pbpErr.Error())
	}

	pb, err := pbp.GetPlaybook()

	if err != nil {
		log.Fatalf("Error getting playbook: %s", err.Error())
	}

	pb.MergeCLIVariables(options.variables)

	sp, spErr := SQLProviderFromOptions(options)

	if spErr != nil {
		log.Fatalf("Could not determine sql source: %s", spErr.Error())
	}

	statuses := run.Run(*pb, sp, options.fromStep, options.dryRun)
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

func PlaybookProviderFromOptions(options Options) (playbook.PlaybookProvider, error) {
	if options.consul != "" {
		return playbook.NewConsulPlaybookProvider(options.consul, options.playbook), nil
	} else if options.playbook != "" {
		return playbook.NewYAMLFilePlaybookProvider(options.playbook), nil
	} else {
		return nil, errors.New("Cannot determine provider for playbook")
	}
}

func SQLProviderFromOptions(options Options) (playbook.SQLProvider, error) {
	if options.consul != "" {
		return playbook.NewConsulSQLProvider(options.consul, options.sqlroot), nil
	} else if options.playbook != "" {
		return playbook.NewFileSQLProvider(options.sqlroot), nil
	} else {
		return nil, errors.New("Cannot determine provider for sql")
	}
}

// Resolve the path to our SQL scripts
func resolveSqlRoot(sqlroot string, playbookPath string, consulAddress string) (string, error) {
	consulErr1 := fmt.Errorf("Cannot use %s option with -consul argument", sqlroot)
	consulErr2 := fmt.Errorf("Cannot use %s option without -consul argument", sqlroot)

	switch sqlroot {
	case SQLROOT_BINARY:
		if consulAddress != "" {
			return "", consulErr1
		}
		return osext.ExecutableFolder()
	case SQLROOT_PLAYBOOK:
		if consulAddress != "" {
			return getAbsConsulPath(playbookPath), nil
		}
		return filepath.Abs(filepath.Dir(playbookPath))
	case SQLROOT_PLAYBOOK_CHILD:
		if consulAddress != "" {
			return playbookPath, nil
		}
		return "", consulErr2
	default:
		return sqlroot, nil
	}
}

// Gets an absolute path for Consul one directory up
func getAbsConsulPath(path string) string {
	strSpl := strings.Split(path, "/")
	trimSpl := strSpl[:len(strSpl)-1]
	return strings.Join(trimSpl, "/")
}
