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

home=${HOME}
aws_dir=${home}/.aws
creds_file=${aws_dir}/credentials



# -----------------------------------------------------------------------------
#  EXECUTION
# -----------------------------------------------------------------------------

mkdir -p ${aws_dir}
touch ${creds_file}

echo "[default]" >> ${creds_file}
echo "aws_access_key_id=some-aws-key" >> ${creds_file}
echo "aws_secret_access_key=some-aws-secret" >> ${creds_file}
