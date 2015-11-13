//
// Copyright (c) 2015 Snowplow Analytics Ltd. All rights reserved.
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
package playbook

import (
	"fmt"
	"github.com/hashicorp/consul/api"
)

// Attempts to return the string value
// of a key stored in a consul server
func GetStringValueFromConsul(address string, key string) (string, error) {
	// Add address to config
	conf := api.DefaultConfig()
	conf.Address = address

	// Connect to consul
	client, err := api.NewClient(conf)
	if err != nil {
		return "", err
	}

	kv := client.KV()

	// Get the KV Pair from consul
	pair, _, err := kv.Get(key, nil)
	if err != nil {
		return "", err
	}

	if pair != nil {
		return byteArrayToString(pair.Value), nil
	} else {
		return "", fmt.Errorf("The key '%s' returned a nil value from the consul server", key)
	}
}

// Converts a byte array into a string
func byteArrayToString(byteArray []byte) string {
	return string(byteArray[:])
}
