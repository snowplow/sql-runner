# SQL Runner

[![Build Status][travis-image]][travis] [![Coveralls][coveralls-image]][coveralls] [![Go Report Card][goreport-image]][goreport] [![Release][release-image]][releases] [![License][license-image]][license]

## Overview

Run playbooks of SQL scripts in series and parallel on Snowflake DB, Amazon Redshift and PostgreSQL.

Used with **[Snowplow][snowplow]** for scheduled SQL-based transformations of event stream data.

|  **[Devops Guide][devops-guide]**     | **[Analysts Guide][analysts-guide]**     | **[Developers Guide][developers-guide]**     |
|:--------------------------------------:|:-----------------------------------------:|:---------------------------------------------:|
|  [![i1][devops-image]][devops-guide] | [![i2][analysts-image]][analysts-guide] | [![i3][developers-image]][developers-guide] |

## Quick start

Assuming [go][go-url], [docker][docker-url] and [docker-compose][docker-compose-url] are installed:

```bash
 host> git clone https://github.com/snowplow/sql-runner
 host> cd sql-runner
 host> make setup-up    # Launches Consul + Postgres for testing
 host> make             # Builds sql-runner binaries
 host> make test        # Runs unit tests

 # DISTRO specifies which binary you want to run integration tests with
 host> DISTRO=darwin make integration
```

_Note_: You will need to ensure that `~/go/bin` is on your PATH for `gox` to work - the underlying tool that we use for building the binaries.

When you are done with testing simply execute `make setup-down` to terminate the docker-compose stack.

To reset the testing resources execute `make setup-reset` which will rebuild the docker containers.  This can be useful if the state of these systems gets out of sync with what the tests expect.

To remove all build files:

```bash
guest> make clean
```

To format the golang code in the source directory:

```bash
guest> make format
```

**Note:** Always run `make format` before submitting any code.

**Note:** The `make test` command also generates a code coverage file which can be found at `build/coverage/coverage.html`.

## How to use?

First either compile the binary from source using the above `make` command or download the published Binary directly from Bintray:

* [Darwin (macOS)](https://dl.bintray.com/snowplow/snowplow-generic/sql_runner_0.9.3_darwin_amd64.zip)
* [Linux](https://dl.bintray.com/snowplow/snowplow-generic/sql_runner_0.9.3_linux_amd64.zip)
* [Windows](https://dl.bintray.com/snowplow/snowplow-generic/sql_runner_0.9.3_windows_amd64.zip)

### CLI Output

```bash
sql-runner version: 0.9.3
Run playbooks of SQL scripts in series and parallel on Redshift and Postgres
Usage:
  -checkLock string
    	Checks whether the lockfile already exists
  -consul string
    	The address of a consul server with playbooks and SQL files stored in KV pairs
  -consulOnlyForLock
    	Will read playbooks locally, but use Consul for locking.
  -deleteLock string
    	Will attempt to delete a lockfile if it exists
  -dryRun
    	Runs through a playbook without executing any of the SQL
  -fillTemplates
    	Will print all queries after templates are filled
  -fromStep string
    	Starts from a given step defined in your playbook
  -help
    	Shows this message
  -lock string
    	Optional argument which checks and sets a lockfile to ensure this run is a singleton. Deletes lock on run completing successfully
  -playbook string
    	Playbook of SQL scripts to execute
  -runQuery string
    	Will run a single query in the playbook
  -showQueryOutput
    	Will print all output from queries
  -softLock string
    	Optional argument, like '-lock' but the lockfile will be deleted even if the run fails
  -sqlroot string
    	Absolute path to SQL scripts. Use PLAYBOOK, BINARY and PLAYBOOK_CHILD for those respective paths (default "PLAYBOOK")
  -var value
    	Variables to be passed to the playbook, in the key=value format
  -version
    	Shows the program version
```

## Copyright and license

SQL Runner is copyright 2015-2021 Snowplow Analytics Ltd.

Licensed under the **[Apache License, Version 2.0][license]** (the "License");
you may not use this software except in compliance with the License.

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

[go-url]: https://golang.org/doc/install
[docker-url]: https://docs.docker.com/get-docker/
[docker-compose-url]: https://docs.docker.com/compose/install/

[travis]: https://travis-ci.org/snowplow/sql-runner
[travis-image]: https://travis-ci.org/snowplow/sql-runner.png?branch=master

[release-image]: https://img.shields.io/github/v/release/snowplow/sql-runner
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

[devops-image]:  https://d3i6fms1cm1j0i.cloudfront.net/github/images/setup.png
[analysts-image]: https://d3i6fms1cm1j0i.cloudfront.net/github/images/techdocs.png
[developers-image]:  https://d3i6fms1cm1j0i.cloudfront.net/github/images/setup.png
