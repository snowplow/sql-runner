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
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
)

func TestMergeCLIVariables(t *testing.T) {
	testCases := []struct {
		Name     string
		Vars     map[string]string
		Expected Playbook
	}{
		{
			Name:     "empty_map",
			Vars:     make(map[string]string),
			Expected: NewPlaybook(),
		},
		{
			Name: "happy_path",
			Vars: map[string]string{
				"a": "A",
				"b": "B",
			},
			Expected: Playbook{
				Variables: map[string]interface{}{
					"a": "A",
					"b": "B",
				},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			pb := NewPlaybook()
			pb.MergeCLIVariables(tt.Vars)

			if !reflect.DeepEqual(pb, tt.Expected) {
				t.Errorf("GOT:\n%s\nEXPECTED:\n%s",
					spew.Sdump(pb),
					spew.Sdump(tt.Expected))
			}
		})
	}
}

func TestValidate(t *testing.T) {
	testCases := []struct {
		Name      string
		Play      Playbook
		IsValid   bool
		ErrString string
	}{
		{
			Name: "nil_targets",
			Play: Playbook{
				Targets: nil,
				Steps:   make([]Step, 1),
			},
			IsValid:   false,
			ErrString: "no targets",
		},
		{
			Name: "zero_targets",
			Play: Playbook{
				Targets: make([]Target, 0),
				Steps:   make([]Step, 1),
			},
			IsValid:   false,
			ErrString: "no targets",
		},
		{
			Name: "nil_steps",
			Play: Playbook{
				Targets: make([]Target, 1),
				Steps:   nil,
			},
			IsValid:   false,
			ErrString: "no steps",
		},
		{
			Name: "zero_steps",
			Play: Playbook{
				Targets: make([]Target, 1),
				Steps:   make([]Step, 0),
			},
			IsValid:   false,
			ErrString: "no steps",
		},
		{
			Name: "happy_path",
			Play: Playbook{
				Targets: make([]Target, 1),
				Steps:   make([]Step, 1),
			},
			IsValid:   true,
			ErrString: "",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			assert := assert.New(t)

			err := tt.Play.Validate()

			if tt.IsValid {
				assert.Nil(err)
			} else {
				if err == nil {
					t.Fatal("got error, expected nil")
				}
				assert.Equal(tt.ErrString, err.Error())
			}
		})
	}
}
