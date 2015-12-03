# SQL Runner

[ ![Build Status] [travis-image] ] [travis] [ ![Release] [release-image] ] [releases] [ ![License] [license-image] ] [license]

## Overview

Run playbooks of SQL scripts in series and parallel on Amazon Redshift and PostgreSQL.

Used with **[Snowplow] [snowplow]** for scheduled SQL-based transformations of event stream data.

## User Quickstart

Assuming you are running on **64bit Linux**:

```bash
> wget http://dl.bintray.com/snowplow/snowplow-generic/sql_runner_0.1.0_linux_amd64.zip
> unzip sql_runner_0.1.0_linux_amd64.zip
> ./sql-runner -usage
```

See the **User Guide** section below for more.

## Developer Quickstart

### Building

Assuming git, **[Vagrant] [vagrant-install]** and **[VirtualBox] [virtualbox-install]** installed:

```bash
 host> git clone https://github.com/snowplow/sql-runner
 host> cd sql-runner
 host> vagrant up && vagrant ssh
guest> cd /opt/gopath/src/github.com/snowplow/sql-runner
guest> godep go build
```

### Testing

Assuming **Building** complete:

```
guest> sudo -u postgres psql
psql# \password
psql# Enter password: postgres
psql# \q
guest> psql -c 'create database sql_runner_tests_1' -U postgres
guest> psql -c 'create database sql_runner_tests_2' -U postgres
guest> ./sql-runner -playbook ./integration-tests/good-postgres.yml -var test_date=`date "+%Y_%m_%d"`
```

### Publishing

Assuming **[Travis] [travis]** is green and versions updated:

```bash
 host> vagrant push
```

This will build an individual artifact for Windows, OSX and Linux all in 64 bit.  All artifacts are stored in the `dist/` directory.

## User guide

### CLI Arguments

There are several command line arguments that can be used:

* `-playbook` : This is a required argument and should point to the playbook you wish to run.
* `-fromStep` : Optional argument which will allow you to start the sql-runner from any step in your playbook.
* `-sqlroot`  : Optional argument to change where we look for the sql statements to run, defaults to the directory of your playbook.
* `-var`      : Optional argument which allows you to pass a dictionary of key-value pairs which will be used to flesh out your templates.
* `-consul`   : Optional argument which allows you to fetch playbooks and SQL files from a Consul server.
* `-dryRun`   : Optional argument which allows you to run through your playbook without executing any SQL against your target(s)

#### More on Consul

Using the `-consul` argument results in the following changes:

* The `-playbook` argument becomes the key that is used to look for the playbook in Consul.
* The `-sqlroot` argument also becomes a key argument for Consul.

If you pass in the default:

* `./sql-runner -consul "localhost:8500" -playbook "sql-runner/playbook/1"`

This results in:

* Looking for your playbook file at this key `sql-runner/playbook/1`
* Expecting all your SQL file keys to begin with `sql-runner/playbook/<SQL path from playbook>`

However as the key here can be used as a both a data and folder node we have added a new sqlroot option:

* `./sql-runner -consul "localhost:8500" -playbook "sql-runner/playbook/1" -sqlroot PLAYBOOK_CHILD`

This results in:

* Looking for your playbook file at this key `sql-runner/playbook/1`
* Expecting all your SQL file keys to begin with `sql-runner/playbook/1/<SQL path from playbook>`
  - The data node is used as a folder node as well.

### Playbooks

A playbook consists of one of more _steps_, each of which consists of one or more _queries_. Steps are run in series, queries are run in parallel within the step. 

Each query contains the path to a _query file_. See **Query files** for details

All steps are applied against all _targets_. All targets are processed in parallel.

For the playbook template see: [config/config.yml.sample] [example-config]

### Query files

A query file contains one or more SQL statements. These are executed "raw" (i.e. not in a transaction) in series by SQL Runner. 

If the query file is flagged as a _template_ in the playbook, then the file is pre-processed as a template before being executed. See **Templates** for details

### Templates

Templates are run through Golang's [text template processor] [go-text-template]. The template processor can access all _variables_ defined in the playbook.

The following custom functions are also supported:

* `nowWithFormat [timeFormat]`: where `timeFormat` is a valid Golang [time format] [go-time-format]
* `systemEnv "ENV_VAR"`: where `ENV_VAR` is a key for a valid environment variable
* `awsEnvCredentials`: supports passing credentials through environment variables, such as `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`
* `awsProfileCredentials`: supports getting credentials from a credentials file, also used by boto/awscli
* `awsEC2RoleCredentials`: supports getting role-based credentials, i.e. getting the automatically generated credentials in EC2 instances
* `awsChainCredentials`: tries to get credentials from each of the three methods above in order, using the first one returned

**Note**: All AWS functions output strings in the Redshift credentials format (`CREDENTIALS 'aws_access_key_id=%s;aws_secret_access_key=%s'`).

For an example query file using templating see: [integration-tests/postgres-sql/good/3.sql] [example-query]

### Failure modes

If a statement fails in a query file, the query will terminate and report failure.

If a query fails, its sibling queries will continue running, but no further steps will run.

Failures in one target do not affect other targets in any way.

## Copyright and license

SQL Runner is copyright 2015 Snowplow Analytics Ltd.

Licensed under the **[Apache License, Version 2.0] [license]** (the "License");
you may not use this software except in compliance with the License.

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

[travis]: https://travis-ci.org/snowplow/sql-runner
[travis-image]: https://travis-ci.org/snowplow/sql-runner.png?branch=master

[release-image]: http://img.shields.io/badge/release-0.4.0-6ad7e5.svg?style=flat
[releases]: https://github.com/snowplow/sql-runner/releases

[license-image]: http://img.shields.io/badge/license-Apache--2-blue.svg?style=flat
[license]: http://www.apache.org/licenses/LICENSE-2.0

[vagrant-install]: http://docs.vagrantup.com/v2/installation/index.html
[virtualbox-install]: https://www.virtualbox.org/wiki/Downloads

[example-config]: https://raw.githubusercontent.com/snowplow/sql-runner/master/config/config.yml.sample
[example-query]: https://raw.githubusercontent.com/snowplow/sql-runner/master/integration-tests/postgres-sql/good/3.sql

[go-text-template]: http://golang.org/pkg/text/template/
[go-time-format]: http://golang.org/pkg/time/#Time.Format

[snowplow]: https://github.com/snowplow/snowplow

