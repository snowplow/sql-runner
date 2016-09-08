-- Test file: 1.sql

DROP SCHEMA IF EXISTS {{.test_schema}} CASCADE;

CREATE SCHEMA {{.test_schema}};

CREATE TABLE {{.test_schema}}.table1 (
  age int
);
