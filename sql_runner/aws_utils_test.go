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
	"os"
	"testing"
)

func TestAwsEnvCredentials(t *testing.T) {
	assert := assert.New(t)

	str, err := awsEnvCredentials()
	assert.NotNil(err)
	assert.NotNil(str)
	assert.Equal("EnvAccessKeyNotFound: AWS_ACCESS_KEY_ID or AWS_ACCESS_KEY not found in environment", err.Error())
	assert.Equal("CREDENTIALS 'aws_access_key_id=;aws_secret_access_key='", str)

	os.Setenv("AWS_ACCESS_KEY_ID", "some-aws-key")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "some-aws-secret")

	str, err = awsEnvCredentials()
	assert.NotNil(str)
	assert.Nil(err)
	assert.Equal("CREDENTIALS 'aws_access_key_id=some-aws-key;aws_secret_access_key=some-aws-secret'", str)
}

func TestAwsProfileCredentials(t *testing.T) {
	assert := assert.New(t)

	str, err := awsProfileCredentials("fake-profile")
	assert.NotNil(err)
	assert.NotNil(str)
	assert.Equal("CREDENTIALS 'aws_access_key_id=;aws_secret_access_key='", str)

	/**
	str, err = awsProfileCredentials("default")
	assert.NotNil(str)
	assert.Nil(err)
	assert.Equal("CREDENTIALS 'aws_access_key_id=some-aws-key;aws_secret_access_key=some-aws-secret'", str)
	*/
}

func TestAwsChainCredentials(t *testing.T) {
	assert := assert.New(t)

	os.Setenv("AWS_ACCESS_KEY_ID", "some-aws-key")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "some-aws-secret")

	str, err := awsChainCredentials("fake-profile")
	assert.NotNil(str)
	assert.Nil(err)
	assert.Equal("CREDENTIALS 'aws_access_key_id=some-aws-key;aws_secret_access_key=some-aws-secret'", str)

	str, err = awsChainCredentials("default")
	assert.NotNil(str)
	assert.Nil(err)
	assert.Equal("CREDENTIALS 'aws_access_key_id=some-aws-key;aws_secret_access_key=some-aws-secret'", str)
}
