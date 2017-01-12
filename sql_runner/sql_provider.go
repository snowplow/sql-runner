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
	"io/ioutil"
	"path"
)

type SQLProvider interface {
	ResolveKey(key string) string
	GetSQL(key string) (string, error)
}

type FileSQLProvider struct {
	rootPath string
}

func NewFileSQLProvider(rootPath string) *FileSQLProvider {
	return &FileSQLProvider{
		rootPath: rootPath,
	}
}

func (p FileSQLProvider) GetSQL(scriptPath string) (string, error) {
	return readScript(p.ResolveKey(scriptPath))
}

func (p FileSQLProvider) ResolveKey(scriptPath string) string {
	return path.Join(p.rootPath, scriptPath)
}

// Reads the file ready for executing
func readScript(file string) (string, error) {
	scriptBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	return string(scriptBytes), nil
}

type ConsulSQLProvider struct {
	consulAddress string
	keyPrefix     string
}

func NewConsulSQLProvider(consulAddress, keyPrefix string) *ConsulSQLProvider {
	return &ConsulSQLProvider{
		consulAddress: consulAddress,
		keyPrefix:     keyPrefix,
	}
}

func (p ConsulSQLProvider) GetSQL(key string) (string, error) {
	return GetStringValueFromConsul(p.consulAddress, p.ResolveKey(key))
}

func (p ConsulSQLProvider) ResolveKey(key string) string {
	return path.Join(p.keyPrefix, key)
}
