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
	"reflect"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
)

func TestTrimToQuery_Valid(t *testing.T) {
	testTargets := []Target{{Name: "test"}}
	testSteps := []Step{
		{
			Name:    "preFoo",
			Queries: []Query{{Name: "bar"}},
		},
		{
			Name:    "foo",
			Queries: []Query{{Name: "bar"}},
		},
		{
			Name:    "postFoo",
			Queries: []Query{{Name: "bar"}},
		},
	}

	testCases := []struct {
		RunQueryArg string
		ExpectedIdx int
	}{
		{
			RunQueryArg: "preFoo::bar",
			ExpectedIdx: 0,
		},
		{
			RunQueryArg: "foo::bar",
			ExpectedIdx: 1,
		},
		{
			RunQueryArg: "postFoo::bar",
			ExpectedIdx: 2,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.RunQueryArg, func(t *testing.T) {
			assert := assert.New(t)
			steps, statuses := trimToQuery(
				testSteps,
				tt.RunQueryArg,
				testTargets)

			assert.Nil(statuses)

			if tt.ExpectedIdx >= len(testSteps) {
				t.Fatal("expected index out of testSteps range")
			}

			i := tt.ExpectedIdx
			expectedSteps := testSteps[i : i+1]
			if !reflect.DeepEqual(steps, testSteps[i:i+1]) {
				t.Errorf("\nGOT:\n%s\nEXPECTED:\n%s",
					spew.Sdump(steps),
					spew.Sdump(expectedSteps))
			}
		})
	}
}

func TestTrimToQuery_Errors(t *testing.T) {
	testTargets := []Target{
		{Name: "a"},
		{Name: "b"},
	}
	testSteps := []Step{
		{
			Name:    "foo",
			Queries: []Query{{Name: "bar"}},
		},
	}

	testCases := []struct {
		Scenario    string
		RunQueryArg string
		ErrorString string
		ErrorExact  bool
	}{
		{
			Scenario:    "missing_delimiter",
			RunQueryArg: "foobar",
			ErrorString: errorRunQueryArgument,
			ErrorExact:  true,
		},
		{
			Scenario:    "missing_delimiter_existing_step_issue210",
			RunQueryArg: "foo",
			ErrorString: errorRunQueryArgument,
			ErrorExact:  true,
		},
		{
			Scenario:    "missing_step",
			RunQueryArg: "::bar",
			ErrorString: errorRunQueryArgument,
			ErrorExact:  true,
		},
		{
			Scenario:    "missing_query",
			RunQueryArg: "foo::",
			ErrorString: errorRunQueryArgument,
			ErrorExact:  true,
		},
		{
			Scenario:    "empty_string",
			RunQueryArg: "",
			ErrorString: errorRunQueryArgument,
			ErrorExact:  true,
		},
		{
			Scenario:    "step_not_found",
			RunQueryArg: "baz::bar",
			ErrorString: errorFromStepNotFound,
			ErrorExact:  false,
		},
		{
			Scenario:    "query_not_found",
			RunQueryArg: "foo::foo",
			ErrorString: errorRunQueryNotFound,
			ErrorExact:  false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.Scenario, func(t *testing.T) {
			assert := assert.New(t)
			steps, statuses := trimToQuery(
				testSteps,
				tt.RunQueryArg,
				testTargets)

			assert.Nil(steps)
			if statuses == nil {
				t.Fatal("unexpected nil statuses")
			}

			if len(statuses) != len(testTargets) {
				t.Fatalf("wrong length of []TargetStatus returned: got %v, expected %v", len(statuses), len(testTargets))
			}

			for i, status := range statuses {
				assert.Equal(status.Name, testTargets[i].Name)

				if status.Errors == nil {
					t.Fatalf("unexpected nil errors in status - got status:\n%s\n", spew.Sdump(status))
				}

				for i, ee := range status.Errors {
					if !checkError(ee, tt.ErrorString, tt.ErrorExact) {
						t.Errorf("error mismatch (at Errors index: %v) in status:\n%s\nEXPECTED (exact: %v) error string: %q\n",
							i,
							spew.Sdump(status),
							tt.ErrorExact,
							tt.ErrorString)
					}
				}
			}
		})
	}
}

// Helper
func checkError(err error, errString string, exact bool) bool {
	if err == nil {
		return false
	}

	if !exact {
		return strings.Contains(err.Error(), errString)
	}

	return err.Error() == errString
}
