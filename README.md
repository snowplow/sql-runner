# SQL Runner

[![Build Status][travis-image]][travis] [![Coveralls][coveralls-image]][coveralls] [![Go Report Card][goreport-image]][goreport] [![Release][release-image]][releases] [![License][license-image]][license]

## Overview

Run playbooks of SQL scripts in series and parallel on Snowflake DB, Amazon Redshift and PostgreSQL.

Used with **[Snowplow][snowplow]** for scheduled SQL-based transformations of event stream data.

|  **[Devops Guide][devops-guide]**     | **[Analysts Guide][analysts-guide]**     | **[Developers Guide][developers-guide]**     |
|:--------------------------------------:|:-----------------------------------------:|:---------------------------------------------:|
|  [![i1][devops-image]][devops-guide] | [![i2][analysts-image]][analysts-guide] | [![i3][developers-image]][developers-guide] |

## Quickstart

Assuming you are running on **64bit Linux**:

```bash
> wget http://dl.bintray.com/snowplow/snowplow-generic/sql_runner_0.6.0_linux_amd64.zip
> unzip sql_runner_0.6.0_linux_amd64.zip
> ./sql-runner -usage
```

## Copyright and license

SQL Runner is copyright 2015-2017 Snowplow Analytics Ltd.

Licensed under the **[Apache License, Version 2.0][license]** (the "License");
you may not use this software except in compliance with the License.

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

[travis]: https://travis-ci.org/snowplow/sql-runner
[travis-image]: https://travis-ci.org/snowplow/sql-runner.png?branch=master

[release-image]: http://img.shields.io/badge/release-0.6.0-6ad7e5.svg?style=flat
[releases]: https://github.com/snowplow/sql-runner/releases

[license-image]: http://img.shields.io/badge/license-Apache--2-blue.svg?style=flat
[license]: http://www.apache.org/licenses/LICENSE-2.0

[coveralls-image]: https://coveralls.io/repos/github/snowplow/sql-runner/badge.svg?branch=master
[coveralls]: https://coveralls.io/github/snowplow/sql-runner?branch=master

[goreport]: https://goreportcard.com/report/github.com/snowplow/sql-runner
[goreport-image]: https://goreportcard.com/badge/github.com/snowplow/sql-runner

[snowplow]: https://github.com/snowplow/snowplow

[analysts-guide]: https://github.com/snowplow/sql-runner/wiki/Guide-for-analysts
[developers-guide]: https://github.com/snowplow/sql-runner/wiki/Guide-for-developers
[devops-guide]: https://github.com/snowplow/sql-runner/wiki/Guide-for-devops

[devops-image]:  http://sauna-github-static.s3-website-us-east-1.amazonaws.com/devops.svg
[analysts-image]: http://sauna-github-static.s3-website-us-east-1.amazonaws.com/analyst.svg
[developers-image]:  http://sauna-github-static.s3-website-us-east-1.amazonaws.com/developer.svg
