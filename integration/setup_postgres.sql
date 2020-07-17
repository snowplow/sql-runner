CREATE USER snowplow WITH PASSWORD 'snowplow';
ALTER ROLE snowplow WITH superuser;
CREATE DATABASE sql_runner_tests_1 OWNER snowplow;
CREATE DATABASE sql_runner_tests_2 OWNER snowplow;
