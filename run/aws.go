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
package run

import (
	"github.com/aws/aws-sdk-go/aws/credentials"
	"fmt"
)

func awsCredentials(creds *credentials.Credentials) (string, error) {
	value, err := creds.Get()
	accessKeyID := value.AccessKeyID
	secretAccessKey := value.SecretAccessKey
	return fmt.Sprintf("CREDENTIALS 'aws_access_key_id=%s;aws_secret_access_key=%s'", accessKeyID, secretAccessKey), err
}

func awsEnvCredentials() (string, error) {
	creds := credentials.NewEnvCredentials()
	return awsCredentials(creds)
}

func awsProfileCredentials(profile string) (string, error) {
	creds := credentials.NewSharedCredentials("", profile)
	return awsCredentials(creds)
}
