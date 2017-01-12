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
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestLockFileFromOptions(t *testing.T) {
	assert := assert.New(t)

	options := Options{
		dryRun: true,
	}
	lockFile, err := LockFileFromOptions(options)
	assert.Nil(lockFile)
	assert.Nil(err)

	options = Options{
		dryRun: false,
	}
	lockFile, err = LockFileFromOptions(options)
	assert.Nil(lockFile)
	assert.Nil(err)

	options = Options{
		dryRun: false,
		lock:   "../dist/lock.lockfile",
	}
	lockFile, err = LockFileFromOptions(options)
	assert.Nil(err)
	assert.False(lockFile.SoftLock)
	assert.Equal("../dist/lock.lockfile", lockFile.Path)

	options = Options{
		dryRun:   false,
		softLock: "../dist/lock.lockfile",
	}
	lockFile, err = LockFileFromOptions(options)
	assert.Nil(err)
	assert.True(lockFile.SoftLock)
	assert.Equal("../dist/lock.lockfile", lockFile.Path)

	options = Options{
		dryRun:    false,
		checkLock: "../dist/lock.lockfile",
	}
	lockFile, err = LockFileFromOptions(options)
	assert.Nil(err)
	assert.False(lockFile.SoftLock)
	assert.Equal("../dist/lock.lockfile", lockFile.Path)

	options = Options{
		dryRun:     false,
		deleteLock: "../dist/lock.lockfile",
	}
	lockFile, err = LockFileFromOptions(options)
	assert.Nil(err)
	assert.False(lockFile.SoftLock)
	assert.Equal("../dist/lock.lockfile", lockFile.Path)
}

func TestResolveSqlRoot(t *testing.T) {
	assert := assert.New(t)

	str, err := resolveSqlRoot(SQLROOT_BINARY, "../integration/resources/good-postgres.yml", "")
	assert.NotNil(str)
	assert.Nil(err)
	str, err = resolveSqlRoot(SQLROOT_BINARY, "../integration/resources/good-postgres.yml", "localhost:8500")
	assert.NotNil(str)
	assert.NotNil(err)
	assert.Equal("", str)
	assert.Equal("Cannot use BINARY option with -consul argument", err.Error())

	str, err = resolveSqlRoot(SQLROOT_PLAYBOOK, "../integration/resources/good-postgres.yml", "")
	assert.NotNil(str)
	assert.Nil(err)
	assert.True(strings.HasSuffix(str, "/integration/resources"))
	str, err = resolveSqlRoot(SQLROOT_PLAYBOOK, "../integration/resources/good-postgres.yml", "localhost:8500")
	assert.NotNil(str)
	assert.Nil(err)
	assert.True(strings.HasSuffix(str, "/integration/resources"))

	str, err = resolveSqlRoot(SQLROOT_PLAYBOOK_CHILD, "../integration/resources/good-postgres.yml", "")
	assert.NotNil(str)
	assert.NotNil(err)
	assert.Equal("", str)
	assert.Equal("Cannot use PLAYBOOK_CHILD option without -consul argument", err.Error())
	str, err = resolveSqlRoot(SQLROOT_PLAYBOOK_CHILD, "../integration/resources/good-postgres.yml", "localhost:8500")
	assert.NotNil(str)
	assert.Nil(err)
	assert.Equal("../integration/resources/good-postgres.yml", str)

	str, err = resolveSqlRoot("random", "../integration/resources/good-postgres.yml", "localhost:8500")
	assert.NotNil(str)
	assert.Nil(err)
	assert.Equal("random", str)
}
