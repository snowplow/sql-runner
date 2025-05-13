// Copyright (c) 2015-2025 Snowplow Analytics Ltd. All rights reserved.
//
// This program is licensed to you under the Apache License Version 2.0,
// and you may not use this file except in compliance with the Apache License Version 2.0.
// You may obtain a copy of the Apache License Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0.
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the Apache License Version 2.0 is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the Apache License Version 2.0 for the specific language governing permissions and limitations there under.
package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePrivateKey(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "snowflake-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	keyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatalf("Failed to marshal private key: %v", err)
	}

	testCases := []struct {
		name        string
		setup       func() string
		passphrase  string
		expectError bool
	}{
		{
			name: "valid unencrypted key",
			setup: func() string {
				path := filepath.Join(tmpDir, "unencrypted.pem")
				block := &pem.Block{
					Type:  "PRIVATE KEY",
					Bytes: keyBytes,
				}
				if err := os.WriteFile(path, pem.EncodeToMemory(block), 0600); err != nil {
					t.Fatalf("Failed to write test file: %v", err)
				}
				return path
			},
			passphrase:  "",
			expectError: false,
		},
		{
			name: "valid encrypted key",
			setup: func() string {
				path := filepath.Join(tmpDir, "encrypted.pem")
				block, err := x509.EncryptPEMBlock(rand.Reader, "PRIVATE KEY", keyBytes, []byte("testpass"), x509.PEMCipherAES256)
				if err != nil {
					t.Fatalf("Failed to encrypt key: %v", err)
				}
				if err := os.WriteFile(path, pem.EncodeToMemory(block), 0600); err != nil {
					t.Fatalf("Failed to write test file: %v", err)
				}
				return path
			},
			passphrase:  "testpass",
			expectError: false,
		},
		{
			name: "wrong passphrase",
			setup: func() string {
				path := filepath.Join(tmpDir, "wrong-pass.pem")
				block, err := x509.EncryptPEMBlock(rand.Reader, "PRIVATE KEY", keyBytes, []byte("testpass"), x509.PEMCipherAES256)
				if err != nil {
					t.Fatalf("Failed to encrypt key: %v", err)
				}
				if err := os.WriteFile(path, pem.EncodeToMemory(block), 0600); err != nil {
					t.Fatalf("Failed to write test file: %v", err)
				}
				return path
			},
			passphrase:  "wrongpass",
			expectError: true,
		},
		{
			name: "non-existent file",
			setup: func() string {
				return filepath.Join(tmpDir, "nonexistent.pem")
			},
			passphrase:  "",
			expectError: true,
		},
		{
			name: "invalid PEM data",
			setup: func() string {
				path := filepath.Join(tmpDir, "invalid.pem")
				if err := os.WriteFile(path, []byte("not a pem file"), 0600); err != nil {
					t.Fatalf("Failed to write test file: %v", err)
				}
				return path
			},
			passphrase:  "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			path := tc.setup()
			key, err := parsePrivateKey(path, tc.passphrase)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, key)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, key)
				assert.IsType(t, &rsa.PrivateKey{}, key)
			}
		})
	}
}
