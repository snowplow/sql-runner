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

func TestNewPostgresTarget_Error(t *testing.T) {
	expectedErr := "missing target connection parameters"
	testCases := []struct {
		Name  string
		Input Target
	}{
		{
			Name: "missing_host",
			Input: Target{
				Port:     "5432",
				Username: "postgres",
				Database: "postgres",
			},
		},
		{
			Name: "missing_port",
			Input: Target{
				Host:     "5432",
				Username: "postgres",
				Database: "postgres",
			},
		},
		{
			Name: "missing_user",
			Input: Target{
				Host:     "localhost",
				Port:     "5432",
				Database: "postgres",
			},
		},
		{
			Name: "missing_database",
			Input: Target{
				Host:     "localhost",
				Port:     "5432",
				Username: "postgres",
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			assert := assert.New(t)

			result, err := NewPostgresTarget(tt.Input)
			assert.Nil(result)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}

			assert.Equal(expectedErr, err.Error())

		})
	}
}

func TestNewPostgresTarget(t *testing.T) {
	testCases := []struct {
		Name  string
		Input Target
	}{
		{
			Name: "happy_path",
			Input: Target{
				Host:     "localhost",
				Port:     "5432",
				Username: "pguser",
				Database: "postgres",
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			assert := assert.New(t)

			result, err := NewPostgresTarget(tt.Input)
			assert.Nil(err)
			if result == nil {
				t.Fatal("unexpected nil result")
			}

			assert.Equal(result.Target, tt.Input)
		})
	}
}
