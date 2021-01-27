-- Test file: 3.sql

CREATE VIEW {{.test_schema}}.view_{{nowWithFormat .timeFormat}} AS
  SELECT CAST(AVG(age) AS SIGNED) AS average_age FROM {{.test_schema}}.table1;