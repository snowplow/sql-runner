CREATE USER 'snowplow'@'localhost' IDENTIFIED BY 'snowplow';
GRANT ALL PRIVILEGES ON * . * TO 'snowplow'@'localhost';
CREATE DATABASE sql_runner_tests_1 OWNER snowplow;
CREATE DATABASE sql_runner_tests_2 OWNER snowplow;
