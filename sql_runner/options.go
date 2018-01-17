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
	"flag"
	"fmt"
	"strings"
)

type CLIVariables map[string]string

// Implement the Value interface
func (i *CLIVariables) String() string {
	return fmt.Sprintf("%s", *i)
}

func (i *CLIVariables) Set(value string) error {
	var split = strings.SplitN(value, "=", 2)
	if len(split) > 1 {
		key := split[0]
		val := split[1]
		(*i)[key] = val
	}
	return nil
}

type Options struct {
	help            bool
	version         bool
	playbook        string
	sqlroot         string
	fromStep        string
	dryRun          bool
	consul          string
	lock            string
	softLock        string
	checkLock       string
	deleteLock      string
	runQuery        string
	continueOnError bool
	variables       CLIVariables
}

func NewOptions() Options {
	return Options{variables: make(map[string]string)}
}

func (o *Options) GetFlagSet() *flag.FlagSet {
	var fs = flag.NewFlagSet("Options", flag.ContinueOnError)

	fs.BoolVar(&(o.help), "help", false, "Shows this message")
	fs.BoolVar(&(o.version), "version", false, "Shows the program version")
	fs.StringVar(&(o.playbook), "playbook", "", "Playbook of SQL scripts to execute")
	fs.StringVar(&(o.sqlroot), "sqlroot", SQLROOT_PLAYBOOK, fmt.Sprintf("Absolute path to SQL scripts. Use %s, %s and %s for those respective paths", SQLROOT_PLAYBOOK, SQLROOT_BINARY, SQLROOT_PLAYBOOK_CHILD))
	fs.Var(&(o.variables), "var", "Variables to be passed to the playbook, in the key=value format")
	fs.StringVar(&(o.fromStep), "fromStep", "", "Starts from a given step defined in your playbook")
	fs.BoolVar(&(o.dryRun), "dryRun", false, "Runs through a playbook without executing any of the SQL")
	fs.BoolVar(&(o.continueOnError), "continueOnError", false, "Continues execution even a step fails, the exit code will still be different than 0")
	fs.StringVar(&(o.consul), "consul", "", "The address of a consul server with playbooks and SQL files stored in KV pairs")
	fs.StringVar(&(o.lock), "lock", "", "Optional argument which checks and sets a lockfile to ensure this run is a singleton. Deletes lock on run completing successfully")
	fs.StringVar(&(o.softLock), "softLock", "", "Optional argument, like '-lock' but the lockfile will be deleted even if the run fails")
	fs.StringVar(&(o.checkLock), "checkLock", "", "Checks whether the lockfile already exists")
	fs.StringVar(&(o.deleteLock), "deleteLock", "", "Will attempt to delete a lockfile if it exists")
	fs.StringVar(&(o.runQuery), "runQuery", "", "Will run a single query in the playbook")
	// TODO: add format flag if/when we support TOML

	return fs
}
