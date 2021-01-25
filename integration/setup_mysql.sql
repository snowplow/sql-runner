CREATE USER 'snowplow'@'%' IDENTIFIED BY 'snowplow';
GRANT ALL PRIVILEGES ON *.* TO 'snowplow'@'%';
CREATE DATABASE sql_runner_tests_1;
CREATE DATABASE sql_runner_tests_2;
