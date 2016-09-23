#!/bin/bash

# Copyright (c) 2015-2016 Snowplow Analytics Ltd. All rights reserved.
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



# -----------------------------------------------------------------------------
#  EXECUTION
# -----------------------------------------------------------------------------

cd ${root}

printf "Setting up environment for integration tests...\n"

sudo sed -i -re 's/^local\s*all\s*postgres\s*peer/local all postgres trust/' /etc/postgresql/9.4/main/pg_hba.conf

sudo service postgresql restart

psql -U postgres -c "alter role postgres password '';"
psql -U postgres -tc "SELECT 1 FROM pg_database WHERE datname = 'sql_runner_tests_1'" | grep -q 1 || psql -U postgres -c "CREATE DATABASE sql_runner_tests_1"
psql -U postgres -tc "SELECT 1 FROM pg_database WHERE datname = 'sql_runner_tests_2'" | grep -q 1 || psql -U postgres -c "CREATE DATABASE sql_runner_tests_2"

${root}/integration/setup_consul.sh
${root}/integration/setup_aws.sh

printf "Ready for integration tests!\n"
