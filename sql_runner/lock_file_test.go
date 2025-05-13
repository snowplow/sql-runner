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
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestInitLockFile_Local tests setting up a lockfile
// on the local file system
func TestInitLockFile_Local(t *testing.T) {
	assert := assert.New(t)

	lockFile, err := InitLockFile("../dist/lock.lockfile", false, "")

	assert.Nil(err)
	assert.Equal("../dist/lock.lockfile", lockFile.Path)
	assert.False(lockFile.SoftLock)
	assert.Equal("", lockFile.ConsulAddress)
	assert.False(lockFile.LockExists())
}

// TestLockUnlockFile_Local asserts that we can
// lock and unlock using a local file server
func TestLockUnlockFile_Local(t *testing.T) {
	assert := assert.New(t)

	lockFile, err := InitLockFile("../dist/lock.lockfile", false, "")
	assert.Nil(err)
	assert.False(lockFile.LockExists())

	err = lockFile.Lock()
	assert.Nil(err)
	assert.True(lockFile.LockExists())

	err = lockFile.Lock()
	assert.NotNil(err)
	assert.Equal("cannot Lock: LockFile is already locked", err.Error())
	assert.True(lockFile.LockExists())

	_, err2 := InitLockFile("../dist/lock.lockfile", false, "")
	assert.NotNil(err2)
	assert.Equal("../dist/lock.lockfile found on start, previous run failed or is ongoing. Cannot start", err2.Error())

	err = lockFile.Unlock()
	assert.Nil(err)
	assert.False(lockFile.LockExists())

	err = lockFile.Unlock()
	assert.NotNil(err)
	assert.Equal("remove ../dist/lock.lockfile: no such file or directory", err.Error())
	assert.False(lockFile.LockExists())
}

// TestInitLockFile_LocalFailure tests setting up a lockfile
// on the local file system that does not exist
func TestInitLockFile_LocalFailure(t *testing.T) {
	assert := assert.New(t)

	lockFile, err := InitLockFile("dist/lock.lockfile", false, "")

	assert.Nil(err)
	assert.Equal("dist/lock.lockfile", lockFile.Path)
	assert.False(lockFile.SoftLock)
	assert.Equal("", lockFile.ConsulAddress)
	assert.False(lockFile.LockExists())
}

// TestLockUnlockFile_LocalFailure asserts that we can
// lock and unlock using a local file server that does not exist
func TestLockUnlockFile_LocalFailure(t *testing.T) {
	assert := assert.New(t)

	lockFile, err := InitLockFile("dist/lock.lockfile", false, "")
	assert.Nil(err)
	assert.False(lockFile.LockExists())

	err = lockFile.Lock()
	assert.NotNil(err)
	assert.Equal("directory for key does not exist", err.Error())
	assert.False(lockFile.LockExists())
}

// TestInitLockFile_Consul tests setting up a lockfile
// on a remote consul server
func TestInitLockFile_Consul(t *testing.T) {
	assert := assert.New(t)

	lockFile, err := InitLockFile("dist/lock.lockfile", false, "localhost:8502")

	assert.Nil(err)
	assert.Equal("dist/lock.lockfile", lockFile.Path)
	assert.False(lockFile.SoftLock)
	assert.Equal("localhost:8502", lockFile.ConsulAddress)
	assert.False(lockFile.LockExists())
}

// TestLockUnlockFile_Consul asserts that we can
// lock and unlock using a consul server
func TestLockUnlockFile_Consul(t *testing.T) {
	assert := assert.New(t)

	lockFile, err := InitLockFile("dist/lock.lockfile", false, "localhost:8502")
	assert.Nil(err)
	assert.False(lockFile.LockExists())

	err = lockFile.Lock()
	assert.Nil(err)
	assert.True(lockFile.LockExists())

	err = lockFile.Lock()
	assert.NotNil(err)
	assert.Equal("cannot Lock: LockFile is already locked", err.Error())
	assert.True(lockFile.LockExists())

	_, err2 := InitLockFile("dist/lock.lockfile", false, "localhost:8502")
	assert.NotNil(err2)
	assert.Equal("dist/lock.lockfile found on start, previous run failed or is ongoing. Cannot start", err2.Error())

	err = lockFile.Unlock()
	assert.Nil(err)
	assert.False(lockFile.LockExists())

	err = lockFile.Unlock()
	assert.Nil(err)
}

// TestInitLockFile_ConsulFailure tests setting up a lockfile
// on a remote consul server that does not exist
func TestInitLockFile_ConsulFailure(t *testing.T) {
	assert := assert.New(t)

	lockFile, err := InitLockFile("dist/lock.lockfile", false, "localhost")

	assert.Nil(err)
	assert.Equal("dist/lock.lockfile", lockFile.Path)
	assert.False(lockFile.SoftLock)
	assert.Equal("localhost", lockFile.ConsulAddress)
	assert.False(lockFile.LockExists())
}

// TestLockUnlockFile_ConsulFailure asserts that we can
// lock and unlock using a consul server that does not exist
func TestLockUnlockFile_ConsulFailure(t *testing.T) {
	assert := assert.New(t)

	lockFile, err := InitLockFile("dist/lock.lockfile", false, "localhost")
	assert.Nil(err)
	assert.False(lockFile.LockExists())

	err = lockFile.Lock()
	assert.NotNil(err)
	assert.False(lockFile.LockExists())

	err = lockFile.Lock()
	assert.NotNil(err)
	assert.False(lockFile.LockExists())

	_, err2 := InitLockFile("dist/lock.lockfile", false, "localhost")
	assert.Nil(err2)
	assert.False(lockFile.LockExists())

	err = lockFile.Unlock()
	assert.NotNil(err)
	assert.False(lockFile.LockExists())

	err = lockFile.Unlock()
	assert.NotNil(err)
	assert.False(lockFile.LockExists())
}
