#!/bin/bash

# Copyright (c) 2015-2021 Snowplow Analytics Ltd. All rights reserved.
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

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
CONSUL_SERVER_URL=http://localhost:8502
ROOT_KEY=${DIR}/resources

# -----------------------------------------------------------------------------
#  EXECUTION
# -----------------------------------------------------------------------------

declare -a consul_keys=(
  "${ROOT_KEY}/good-postgres.yml"
  "${ROOT_KEY}/good-mysql.yml"
  "${ROOT_KEY}/postgres-sql/bad/1.sql"
  "${ROOT_KEY}/postgres-sql/good/1.sql"
  "${ROOT_KEY}/postgres-sql/good/2a.sql"
  "${ROOT_KEY}/postgres-sql/good/2b.sql"
  "${ROOT_KEY}/postgres-sql/good/3.sql"
  "${ROOT_KEY}/postgres-sql/good/assert.sql"
  "${ROOT_KEY}/postgres-sql/good/output.sql"
  "${ROOT_KEY}/mysql-sql-good/1.sql"
  "${ROOT_KEY}/mysql-sql-good/2b.sql"
  "${ROOT_KEY}/mysql-sql-good/3.sql"
  "${ROOT_KEY}/mysql-sql-good/assert.sql"
)

echo " --- Stubbing Consul key values for integration tests --- "

for i in "${!consul_keys[@]}"
  do
      :
      key=${consul_keys[$i]}
      value=`cat ${key}`
      res=`curl -s -XPUT -d "${value}" ${CONSUL_SERVER_URL}/v1/kv${key}`
      echo "PUT result for key ${key}: ${res}"
  done

echo " --- Done! --- "
