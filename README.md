# SQL Runner

[ ![Build Status] [travis-image] ] [travis] [ ![Release] [release-image] ] [releases] [ ![License] [license-image] ] [license]


Run playbooks of SQL scripts in series and parallel on Amazon Redshift and PostgreSQL.

Used with **[Snowplow] [snowplow]** for scheduled SQL-based transformations of event stream data.

## Find out more

- [User Quickstart](https://github.com/snowplow/sql-runner/wiki#user-quickstart)
- [Developer Quickstart](https://github.com/snowplow/sql-runner/wiki#developer-quickstart)
- [User guide](https://github.com/snowplow/sql-runner/wiki#user-guide)

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

