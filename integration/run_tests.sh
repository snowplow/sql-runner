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

root=${DIR}/../
bin_path=${DIR}/../build/output/${DISTRO}/sql-runner
consul_server_uri=localhost:8502
root_key=${DIR}/resources
assert_counter=0

# -----------------------------------------------------------------------------
#  FUNCTIONS & PROCEDURES
# -----------------------------------------------------------------------------

# Similar to Perl die
function die() {
   echo "$@" 1>&2 ; exit 1;
}

# Is passed an exit code and a command and
# will then assert that the exit code matches.
#
# Parameters:
# 1. exit_code
# 2. command
function assert_ExitCodeForCommand() {
   [ "$#" -eq 2 ] || die "2 arguments required, $# provided"
   local __exit_code="$1"
   local __command="$2"
   let "assert_counter+=1"

   printf "RUNNING: Assertion ${assert_counter}:\n - ${__command}\n\n"

   set +e
   eval ${__command}
   retval=`echo $?`
   set -e

   if [ ${retval} -eq ${__exit_code} ] ; then
      printf "\nSUCCESS: Test finished with exit code ${__exit_code}\n\n"
   else
      printf "\nFAIL: Expected exit code ${__exit_code} got ${retval}\n\n"
      exit 1
   fi
}

# -----------------------------------------------------------------------------
#  TEST EXECUTION
# -----------------------------------------------------------------------------

cd ${root}

printf "==========================================================\n"
printf " RUNNING INTEGRATION TESTS\n"
printf "==========================================================\n\n"

# Test: Invalid playbook should return exit code 7
assert_ExitCodeForCommand "7" "${bin_path} -playbook ${root_key}/bad-mixed.yml -verbosity 1"

# Test: Valid playbook with invalid query should return exit code 6
assert_ExitCodeForCommand "6" "${bin_path} -playbook ${root_key}/good-postgres.yml -var test_date=`date "+%Y_%m_%d"`"
assert_ExitCodeForCommand "6" "${bin_path} -playbook ${root_key}/good-mysql.yml -var test_date=`date "+%Y_%m_%d"`"

# Test: Valid playbook which attempts to lock but fails should return exit code 1
assert_ExitCodeForCommand "1" "${bin_path} -playbook ${root_key}/good-postgres.yml -var test_date=`date "+%Y_%m_%d"` -fromStep \"Create schema and table\" -lock ${root}/dist/locks/integration/1"
assert_ExitCodeForCommand "1" "${bin_path} -playbook ${root_key}/good-postgres.yml -var test_date=`date "+%Y_%m_%d"` -fromStep \"Create schema and table\" -lock /locks/integration/1 -consul ${consul_server_uri}"
assert_ExitCodeForCommand "1" "${bin_path} -playbook ${root_key}/good-postgres.yml -var test_date=`date "+%Y_%m_%d"` -fromStep \"Create schema and table\" -softLock ${root}/dist/locks/integration/1"
assert_ExitCodeForCommand "1" "${bin_path} -playbook ${root_key}/good-postgres.yml -var test_date=`date "+%Y_%m_%d"` -fromStep \"Create schema and table\" -softLock /locks/integration/1 -consul ${consul_server_uri}"

# Case 8
assert_ExitCodeForCommand "1" "${bin_path} -playbook ${root_key}/good-mysql.yml -var test_date=`date "+%Y_%m_%d"` -fromStep \"Create schema and table\" -lock ${root}/dist/locks/integration/1"
assert_ExitCodeForCommand "1" "${bin_path} -playbook ${root_key}/good-mysql.yml -var test_date=`date "+%Y_%m_%d"` -fromStep \"Create schema and table\" -lock /locks/integration/1 -consul ${consul_server_uri}"
assert_ExitCodeForCommand "1" "${bin_path} -playbook ${root_key}/good-mysql.yml -var test_date=`date "+%Y_%m_%d"` -fromStep \"Create schema and table\" -softLock ${root}/dist/locks/integration/1"
assert_ExitCodeForCommand "1" "${bin_path} -playbook ${root_key}/good-mysql.yml -var test_date=`date "+%Y_%m_%d"` -fromStep \"Create schema and table\" -softLock /locks/integration/1 -consul ${consul_server_uri}"

# Test: Checking for a lock that does not exist should return exit code 0
assert_ExitCodeForCommand "0" "${bin_path} -checkLock ${root}/dist/locks/integration/1"
assert_ExitCodeForCommand "0" "${bin_path} -checkLock locks/integration/1 -consul ${consul_server_uri}"

# Test: Deleting a lock which does not exist should return exit code 1
assert_ExitCodeForCommand "1" "${bin_path} -deleteLock ${root}/dist/locks/integration/1"
assert_ExitCodeForCommand "1" "${bin_path} -deleteLock locks/integration/1 -consul ${consul_server_uri}"

# 16 - Test: Valid playbook which creates a hard-lock and then fails SHOULD leave the lock around afterwards
assert_ExitCodeForCommand "6" "${bin_path} -playbook ${root_key}/good-postgres.yml -var test_date=`date "+%Y_%m_%d"` -lock locks/integration/1 -consul ${consul_server_uri}"
assert_ExitCodeForCommand "3" "${bin_path} -playbook ${root_key}/good-postgres.yml -var test_date=`date "+%Y_%m_%d"` -lock locks/integration/1 -consul ${consul_server_uri}"
assert_ExitCodeForCommand "3" "${bin_path} -checkLock locks/integration/1 -consul ${consul_server_uri} -verbosity 0"
assert_ExitCodeForCommand "0" "${bin_path} -deleteLock locks/integration/1 -consul ${consul_server_uri} -verbosity 0"
assert_ExitCodeForCommand "6" "${bin_path} -playbook ${root_key}/good-postgres.yml -var test_date=`date "+%Y_%m_%d"` -lock ${root}/dist/integration-lock"
assert_ExitCodeForCommand "3" "${bin_path} -playbook ${root_key}/good-postgres.yml -var test_date=`date "+%Y_%m_%d"` -lock ${root}/dist/integration-lock"
assert_ExitCodeForCommand "3" "${bin_path} -checkLock ${root}/dist/integration-lock -verbosity 0"
assert_ExitCodeForCommand "0" "${bin_path} -deleteLock ${root}/dist/integration-lock -verbosity 0"

# 24 - Test: MySQL Valid playbook which creates a hard-lock and then fails SHOULD leave the lock around afterwards
assert_ExitCodeForCommand "6" "${bin_path} -playbook ${root_key}/good-mysql.yml -var test_date=`date "+%Y_%m_%d"` -lock locks/integration/1 -consul ${consul_server_uri} -verbosity 1"
assert_ExitCodeForCommand "3" "${bin_path} -playbook ${root_key}/good-mysql.yml -var test_date=`date "+%Y_%m_%d"` -lock locks/integration/1 -consul ${consul_server_uri} -verbosity 1"
assert_ExitCodeForCommand "3" "${bin_path} -checkLock locks/integration/1 -consul ${consul_server_uri} -verbosity 0"
assert_ExitCodeForCommand "0" "${bin_path} -deleteLock locks/integration/1 -consul ${consul_server_uri} -verbosity 0"
assert_ExitCodeForCommand "6" "${bin_path} -playbook ${root_key}/good-mysql.yml -var test_date=`date "+%Y_%m_%d"` -lock ${root}/dist/integration-lock -verbosity 1"
assert_ExitCodeForCommand "3" "${bin_path} -playbook ${root_key}/good-mysql.yml -var test_date=`date "+%Y_%m_%d"` -lock ${root}/dist/integration-lock -verbosity 1"
assert_ExitCodeForCommand "3" "${bin_path} -checkLock ${root}/dist/integration-lock -verbosity 1"
assert_ExitCodeForCommand "0" "${bin_path} -deleteLock ${root}/dist/integration-lock -verbosity 1"

# 32 - Test: Valid playbook which creates a soft-lock and then fails SHOULD NOT leave the lock around afterwards
assert_ExitCodeForCommand "6" "${bin_path} -playbook ${root_key}/good-postgres.yml -var test_date=`date "+%Y_%m_%d"` -softLock locks/integration/1 -consul ${consul_server_uri}"
assert_ExitCodeForCommand "0" "${bin_path} -checkLock locks/integration/1 -consul ${consul_server_uri} -verbosity 0"
assert_ExitCodeForCommand "1" "${bin_path} -deleteLock locks/integration/1 -consul ${consul_server_uri} -verbosity 1"
assert_ExitCodeForCommand "6" "${bin_path} -playbook ${root_key}/good-postgres.yml -var test_date=`date "+%Y_%m_%d"` -softLock ${root}/dist/integration-lock"
assert_ExitCodeForCommand "0" "${bin_path} -checkLock ${root}/dist/integration-lock -verbosity 0"
assert_ExitCodeForCommand "1" "${bin_path} -deleteLock ${root}/dist/integration-lock -verbosity 1"

# 38 - Test: MySQL Valid playbook which creates a soft-lock and then fails SHOULD NOT leave the lock around afterwards
assert_ExitCodeForCommand "6" "${bin_path} -playbook ${root_key}/good-mysql.yml -var test_date=`date "+%Y_%m_%d"` -softLock locks/integration/1 -consul ${consul_server_uri}"
assert_ExitCodeForCommand "0" "${bin_path} -checkLock locks/integration/1 -consul ${consul_server_uri}"
assert_ExitCodeForCommand "1" "${bin_path} -deleteLock locks/integration/1 -consul ${consul_server_uri}"
assert_ExitCodeForCommand "6" "${bin_path} -playbook ${root_key}/good-mysql.yml -var test_date=`date "+%Y_%m_%d"` -softLock ${root}/dist/integration-lock"
assert_ExitCodeForCommand "0" "${bin_path} -checkLock ${root}/dist/integration-lock"
assert_ExitCodeForCommand "1" "${bin_path} -deleteLock ${root}/dist/integration-lock"

# 44 - Test: Valid playbook which creates a hard/soft-lock and then succeeds SHOULD NOT leave the lock around afterwards
assert_ExitCodeForCommand "0" "${bin_path} -playbook ${root_key}/good-postgres.yml -var test_date=`date "+%Y_%m_%d"` -fromStep \"Create schema and table\" -lock locks/integration/1 -consul ${consul_server_uri}"
assert_ExitCodeForCommand "0" "${bin_path} -checkLock locks/integration/1 -consul ${consul_server_uri}"
assert_ExitCodeForCommand "0" "${bin_path} -playbook ${root_key}/good-postgres.yml -var test_date=`date "+%Y_%m_%d"` -fromStep \"Create schema and table\" -lock ${root}/dist/integration-lock -verbosity 0"
assert_ExitCodeForCommand "0" "${bin_path} -checkLock ${root}/dist/integration-lock -verbosity 0"
assert_ExitCodeForCommand "0" "${bin_path} -playbook ${root_key}/good-postgres.yml -var test_date=`date "+%Y_%m_%d"` -fromStep \"Create schema and table\" -softLock locks/integration/1 -consul ${consul_server_uri} -verbosity 0"
assert_ExitCodeForCommand "0" "${bin_path} -checkLock locks/integration/1 -consul ${consul_server_uri} -verbosity 0"
assert_ExitCodeForCommand "0" "${bin_path} -playbook ${root_key}/good-postgres.yml -var test_date=`date "+%Y_%m_%d"` -fromStep \"Create schema and table\" -softLock ${root}/dist/integration-lock -verbosity 0"
assert_ExitCodeForCommand "0" "${bin_path} -checkLock ${root}/dist/integration-lock -verbosity 0"

# 52 - Test: MySQL Valid playbook which creates a hard/soft-lock and then succeeds SHOULD NOT leave the lock around afterwards
# BREAK
assert_ExitCodeForCommand "0" "${bin_path} -playbook ${root_key}/good-mysql.yml -var test_date=`date "+%Y_%m_%d"` -fromStep \"Create database and table\" -lock locks/integration/2 -consul ${consul_server_uri}"
assert_ExitCodeForCommand "0" "${bin_path} -checkLock locks/integration/2 -consul ${consul_server_uri}"
assert_ExitCodeForCommand "0" "${bin_path} -playbook ${root_key}/good-mysql.yml -var test_date=`date "+%Y_%m_%d"` -fromStep \"Create database and table\" -lock ${root}/dist/integration-lock"
assert_ExitCodeForCommand "0" "${bin_path} -checkLock ${root}/dist/integration-lock"
assert_ExitCodeForCommand "0" "${bin_path} -playbook ${root_key}/good-mysql.yml -var test_date=`date "+%Y_%m_%d"` -fromStep \"Create database and table\" -softLock locks/integration/2 -consul ${consul_server_uri}"
assert_ExitCodeForCommand "0" "${bin_path} -checkLock locks/integration/2 -consul ${consul_server_uri}"
assert_ExitCodeForCommand "0" "${bin_path} -playbook ${root_key}/good-mysql.yml -var test_date=`date "+%Y_%m_%d"` -fromStep \"Create database and table\" -softLock ${root}/dist/integration-lock"
assert_ExitCodeForCommand "0" "${bin_path} -checkLock ${root}/dist/integration-lock"

# Test: Invalid playbook which creates a hard/soft-lock but is run using -dryRun should return exit code 0
assert_ExitCodeForCommand "5" "${bin_path} -playbook ${root_key}/bad-mixed.yml -lock ${root}/dist/integration-lock -dryRun"
assert_ExitCodeForCommand "0" "${bin_path} -playbook ${root_key}/good-postgres.yml -var test_date=`date "+%Y_%m_%d"` -lock ${root}/dist/integration-lock -dryRun"

# Test: MySQL Invalid playbook which creates a hard/soft-lock but is run using -dryRun should return exit code 0
assert_ExitCodeForCommand "5" "${bin_path} -playbook ${root_key}/bad-mixed.yml -lock ${root}/dist/integration-lock -dryRun"
assert_ExitCodeForCommand "0" "${bin_path} -playbook ${root_key}/good-mysql.yml -var test_date=`date "+%Y_%m_%d"` -lock ${root}/dist/integration-lock -dryRun"

# Test: Valid playbook outputs proper results from playbooks using -showQueryOutput
assert_ExitCodeForCommand "6" "${bin_path} -showQueryOutput -playbook ${root_key}/good-postgres.yml"

# Test: MySQL Valid playbook outputs proper results from playbooks using -showQueryOutput
assert_ExitCodeForCommand "6" "${bin_path} -showQueryOutput -playbook ${root_key}/good-mysql.yml"

# Test: Valid playbook which uses playbook template variables
assert_ExitCodeForCommand "6" "${bin_path} -playbook ${root_key}/good-postgres-with-template.yml -var password=,host=localhost"
assert_ExitCodeForCommand "6" "${bin_path} -playbook ${root_key}/good-postgres-with-template.yml"
assert_ExitCodeForCommand "0" "${bin_path} -playbook ${root_key}/good-postgres-with-template.yml -var username=postgres,password=,host=localhost"

# Test: Truncated steps field in playbook should return exit code 8
assert_ExitCodeForCommand "8" "${bin_path} -playbook ${root_key}/good-postgres-truncated.yml -lock ${root}/dist/integration-lock"

# Test: fillTemplate option should return exit code 8
assert_ExitCodeForCommand "8" "${bin_path} -fillTemplates -playbook ${root_key}/good-postgres-with-template.yml -var username=postgres,password=,host=localhost"

printf "==========================================================\n"
printf " INTEGRATION TESTS SUCCESSFUL\n"
printf "==========================================================\n"
