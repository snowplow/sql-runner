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
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPutGetDelStringValueFromConsul_Failure(t *testing.T) {
	assert := assert.New(t)

	err := PutStringValueToConsul("localhost", "somekey", "somevalue")
	assert.NotNil(err)

	str, err := GetStringValueFromConsul("localhost", "somekey")
	assert.Equal("", str)
	assert.NotNil(err)

	err = DeleteValueFromConsul("localhost", "somekey")
	assert.NotNil(err)
}

func TestPutGetDelStringValueFromConsul_Success(t *testing.T) {
	assert := assert.New(t)

	err := PutStringValueToConsul("localhost:8502", "somekey", "somevalue")
	assert.Nil(err)

	str, err := GetStringValueFromConsul("localhost:8502", "somekey")
	assert.Nil(err)
	assert.NotNil(str)
	assert.Equal("somevalue", str)

	err = DeleteValueFromConsul("localhost:8502", "somekey")
	assert.Nil(err)
}
