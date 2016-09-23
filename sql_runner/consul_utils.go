//
// Copyright (c) 2015-2016 Snowplow Analytics Ltd. All rights reserved.
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
	"github.com/hashicorp/consul/api"
)

func GetConsulClient(address string) (*api.Client, error) {
	// Add address to config
	conf := api.DefaultConfig()
	conf.Address = address

	// Connect to consul
	return api.NewClient(conf)
}

// GetBytesFromConsul attempts to return the bytes
// of a key stored in a Consul server
func GetBytesFromConsul(address string, key string) ([]byte, error) {
	client, _ := GetConsulClient(address)
	kv := client.KV()

	// Get the KV Pair from consul
	pair, _, err := kv.Get(key, nil)
	if err != nil {
		return nil, err
	}

	if pair != nil {
		return pair.Value, nil
	} else {
		return nil, fmt.Errorf("The key '%s' returned a nil value from the consul server", key)
	}
}

// GetStringValueFromConsul attempts to return
// the string value of a key stored in a Consul server
func GetStringValueFromConsul(address string, key string) (string, error) {
	bytes, err := GetBytesFromConsul(address, key)

	if err != nil {
		return "", err
	} else {
		return string(bytes), nil
	}
}

// PutBytesToConsul attempts to push a new
// KV pair to a Consul Server
func PutBytesToConsul(address string, key string, value []byte) error {
	client, _ := GetConsulClient(address)
	kv := client.KV()

	// Put a new KV pair to consul
	p := &api.KVPair{Key: key, Value: value}
	_, err := kv.Put(p, nil)
	return err
}

// PutStringValueToConsul attempts to push a new
// KV pair to a Consul Server
func PutStringValueToConsul(address string, key string, value string) error {
	return PutBytesToConsul(address, key, []byte(value))
}

// DeleteValueFromConsul attempts to delete a
// KV pair from a Consul Server
func DeleteValueFromConsul(address string, key string) error {
	client, _ := GetConsulClient(address)
	kv := client.KV()

	// Delete the KV pair
	_, err := kv.Delete(key, nil)
	return err
}
