//
// Copyright (c) 2015-2022 Snowplow Analytics Ltd. All rights reserved.
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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadLocalFile(t *testing.T) {
	assert := assert.New(t)

	bytes, err := loadLocalFile("/this/path/does/not/exist")
	assert.Nil(bytes)
	assert.NotNil(err)
	assert.Equal("open /this/path/does/not/exist: no such file or directory", err.Error())

	bytes, err = loadLocalFile("../VERSION")
	assert.NotNil(bytes)
	assert.Nil(err)
	assert.Equal(cliVersion, string(bytes))
}
