#!/bin/bash

# Copyright (c) 2015-2017 Snowplow Analytics Ltd. All rights reserved.
#
# This program is licensed to you under the Apache License Version 2.0,
# and you may not use this file except in compliance with the Apache License Version 2.0.
# You may obtain a copy of the Apache License Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0.
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the Apache License Version 2.0 is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the Apache License Version 2.0 for the specific language governing permissions and limitations there under.

set -e



# -----------------------------------------------------------------------------
#  CONSTANTS
# -----------------------------------------------------------------------------

root=${TRAVIS_BUILD_DIR}
dist_dir=${root}/dist
consul_server_url=http://localhost:8500
consul_log_path=${dist_dir}/consul.log
root_key=${root}/integration/resources



# -----------------------------------------------------------------------------
#  EXECUTION
# -----------------------------------------------------------------------------

cd ${dist_dir}

wget -N 'https://releases.hashicorp.com/consul/0.7.2/consul_0.7.2_linux_amd64.zip'
unzip -o "consul_0.7.2_linux_amd64.zip"
./consul --version
./consul agent -server -bootstrap-expect 1 -data-dir /tmp/consul >> ${consul_log_path} 2>&1 & 
sleep 5

declare -a consul_keys=( 
  "${root_key}/good-postgres.yml"
  "${root_key}/postgres-sql/bad/1.sql"
  "${root_key}/postgres-sql/good/1.sql"
  "${root_key}/postgres-sql/good/2a.sql"
  "${root_key}/postgres-sql/good/2b.sql"
  "${root_key}/postgres-sql/good/3.sql"
  "${root_key}/postgres-sql/good/assert.sql"
  "${root_key}/postgres-sql/good/output.sql"
)

for i in "${!consul_keys[@]}"
  do 
      :
      key=${consul_keys[$i]}
      value=`cat ${key}`
      res=`curl -s -XPUT -d "${value}" ${consul_server_url}/v1/kv${key}`
      echo "PUT result for key ${key}: ${res}"
  done
