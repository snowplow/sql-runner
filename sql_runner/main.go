//
// Copyright (c) 2015-2020 Snowplow Analytics Ltd. All rights reserved.
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
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	CLI_NAME        = "sql-runner"
	CLI_DESCRIPTION = `Run playbooks of SQL scripts in series and parallel on Redshift and Postgres`
	CLI_VERSION     = "0.9.2"

	SQLROOT_BINARY         = "BINARY"
	SQLROOT_PLAYBOOK       = "PLAYBOOK"
	SQLROOT_PLAYBOOK_CHILD = "PLAYBOOK_CHILD"
)

// main is the entry point for the application
func main() {

	options := processFlags()

	lockFile, lockErr := LockFileFromOptions(options)
	if lockErr != nil {
		log.Printf("Error: %s", lockErr.Error())
		os.Exit(3)
	}

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

	// Lock it up...
	if lockFile != nil {
		lockErr2 := lockFile.Lock()
		if lockErr2 != nil {
			log.Fatalf("Error making lock: %s", lockErr2.Error())
		}
	}

	statuses := Run(*pb, sp, options.fromStep, options.runQuery, options.dryRun, options.fillTemplates, options.showQueryOutput)
	code, message := review(statuses)

	// Unlock on success and soft-lock
	if lockFile != nil {
		if code == 0 || code == 8 || lockFile.SoftLock {
			lockFile.Unlock()
		}
	}

	log.Printf(message)
	os.Exit(code)
}

// processFlags parses the arguments provided to
// the main function.
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

	if options.checkLock != "" {
		lockFile, lockErr := LockFileFromOptions(options)
		if lockErr != nil {
			log.Printf("Error: %s found, previous run failed or is ongoing", lockFile.Path)
			os.Exit(3)
		} else {
			log.Printf("Success: %s does not exist", lockFile.Path)
			os.Exit(0)
		}
	}

	if options.deleteLock != "" {
		lockFile, lockErr := LockFileFromOptions(options)
		if lockErr != nil {
			unlockErr := lockFile.Unlock()
			if unlockErr != nil {
				log.Printf("Error: %s found but could not delete: %s", lockFile.Path, unlockErr.Error())
				os.Exit(1)
			} else {
				log.Printf("Success: %s found and deleted", lockFile.Path)
				os.Exit(0)
			}
		} else {
			log.Printf("Error: %s does not exist, nothing to delete", lockFile.Path)
			os.Exit(1)
		}
	}

	if options.playbook == "" {
		fmt.Println("required flag not defined: -playbook")
		os.Exit(2)
	}

	sr, err := resolveSqlRoot(options.sqlroot, options.playbook, options.consul, options.consulOnlyForLock)
	if err != nil {
		fmt.Printf("Error resolving -sqlroot: %s\n%s\n", options.sqlroot, err)
		os.Exit(2)
	}
	options.sqlroot = sr // Yech, mutate in place

	return options
}

// --- Options resolvers

// PlaybookProviderFromOptions returns a provider of the Playbook
// based on flags passed in
func PlaybookProviderFromOptions(options Options) (PlaybookProvider, error) {
	if options.consulOnlyForLock {
		return NewYAMLFilePlaybookProvider(options.playbook, options.variables), nil
	} else if options.consul != "" {
		return NewConsulPlaybookProvider(options.consul, options.playbook, options.variables), nil
	} else if options.playbook != "" {
		return NewYAMLFilePlaybookProvider(options.playbook, options.variables), nil
	} else {
		return nil, errors.New("Cannot determine provider for playbook")
	}
}

// SQLProviderFromOptions returns a provider of SQL files
// based on flags passed in
func SQLProviderFromOptions(options Options) (SQLProvider, error) {
	if options.consulOnlyForLock {
		return NewFileSQLProvider(options.sqlroot), nil
	} else if options.consul != "" {
		return NewConsulSQLProvider(options.consul, options.sqlroot), nil
	} else if options.playbook != "" {
		return NewFileSQLProvider(options.sqlroot), nil
	} else {
		return nil, errors.New("Cannot determine provider for sql")
	}
}

// LockFileFromOptions will check if a LockFile already
// exists and will then either:
// 1. Raise an error
// 2. Set a new lock
func LockFileFromOptions(options Options) (*LockFile, error) {

	// Do nothing if dry-run
	if options.dryRun == true {
		return nil, nil
	}

	var lockPath string
	var isSoftLock bool

	if options.lock != "" {
		lockPath = options.lock
		isSoftLock = false
	} else if options.softLock != "" {
		lockPath = options.softLock
		isSoftLock = true
	} else if options.checkLock != "" {
		lockPath = options.checkLock
		isSoftLock = false
	} else if options.deleteLock != "" {
		lockPath = options.deleteLock
		isSoftLock = false
	} else {
		// no-op
		return nil, nil
	}

	lockFile, err := InitLockFile(lockPath, isSoftLock, options.consul)

	return &lockFile, err
}

// --- SQLRoot resolvers

// resolveSqlRoot returns the path to our SQL scripts
func resolveSqlRoot(sqlroot string, playbookPath string, consulAddress string, consulOnlyForLock bool) (string, error) {
	consulErr1 := fmt.Errorf("Cannot use %s option with -consul argument", sqlroot)
	consulErr2 := fmt.Errorf("Cannot use %s option without -consul argument", sqlroot)
	consulErr3 := fmt.Errorf("Cannot use %s option with -consulOnlyForLock argument", sqlroot)

	if consulOnlyForLock {
		switch sqlroot {
		case SQLROOT_BINARY:
			return osext.ExecutableFolder()
		case SQLROOT_PLAYBOOK:
			return filepath.Abs(filepath.Dir(playbookPath))
		case SQLROOT_PLAYBOOK_CHILD:
			return "", consulErr3
		default:
			return sqlroot, nil
		}
	}

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

// getAbsConsulPath returns an absolute path for Consul
// one directory up
func getAbsConsulPath(path string) string {
	strSpl := strings.Split(path, "/")
	trimSpl := strSpl[:len(strSpl)-1]
	return strings.Join(trimSpl, "/")
}
