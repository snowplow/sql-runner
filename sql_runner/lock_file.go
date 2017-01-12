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
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

type LockFile struct {
	Path          string
	SoftLock      bool
	ConsulAddress string
	locked        bool
}

// InitLockFile creates a LockFile object
// which is used to ensures jobs can not run
// at the same time.
func InitLockFile(path string, softLock bool, consulAddress string) (LockFile, error) {
	lockFile := LockFile{
		Path:          path,
		SoftLock:      softLock,
		ConsulAddress: consulAddress,
		locked:        false,
	}

	if lockFile.LockExists() {
		return lockFile, fmt.Errorf("%s found on start, previous run failed or is ongoing. Cannot start", path)
	} else {
		return lockFile, nil
	}
}

// Lock creates a new lock file or kv entry
func (lf *LockFile) Lock() error {
	if lf.locked == true {
		return fmt.Errorf("LockFile is already locked!")
	}

	value := time.Now().UTC().Format("2006-01-02T15:04:05-0700")

	log.Printf("Checking and setting the lockfile at this key '%s'", lf.Path)

	if lf.ConsulAddress == "" {
		// Check if dir exists
		dirStr := filepath.Dir(lf.Path)
		if _, err := os.Stat(dirStr); os.IsNotExist(err) {
			return fmt.Errorf("directory for key does not exist")
		}

		// Create the file
		f, err := os.OpenFile(lf.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		defer f.Close()
		if err != nil {
			return err
		}

		// Write a line to it
		_, err = f.WriteString(value)
		if err != nil {
			return err
		}

		lf.locked = true
		return nil
	} else {
		// Create the KV pair
		err := PutStringValueToConsul(lf.ConsulAddress, lf.Path, value)
		if err != nil {
			return err
		}

		lf.locked = true
		return nil
	}
}

// Unlock deletes the lock or kv entry
func (lf *LockFile) Unlock() error {

	log.Printf("Deleting lockfile at this key '%s'", lf.Path)

	if lf.ConsulAddress == "" {
		// Delete the file
		err := os.Remove(lf.Path)
		if err != nil {
			return err
		}

		lf.locked = false
		return nil
	} else {
		// Delete the KV pair
		err := DeleteValueFromConsul(lf.ConsulAddress, lf.Path)
		if err != nil {
			return err
		}

		lf.locked = false
		return nil
	}
}

// LockExists checks if the lock file
// exists already
func (lf *LockFile) LockExists() bool {
	if lf.ConsulAddress == "" {
		if _, err := os.Stat(lf.Path); os.IsNotExist(err) {
			return false
		} else {
			return true
		}
	} else {
		value, err := GetStringValueFromConsul(lf.ConsulAddress, lf.Path)

		if err != nil && value == "" {
			return false
		} else {
			return true
		}
	}
}
