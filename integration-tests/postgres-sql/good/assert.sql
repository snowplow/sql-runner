-- Test file: assert.sql

CREATE OR REPLACE PROCEDURAL LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION {{.test_schema}}.assert_average_age() RETURNS VOID AS $$
DECLARE
  expected_average_age CONSTANT integer := 25;
BEGIN
  {{/* Update view names to today's date or Travis will error */}}
  IF (SELECT average_age <> expected_average_age FROM {{.test_schema}}.view_2015_09_13) THEN
    RAISE EXCEPTION 'Average_age % does not match expected age %',
    	(SELECT average_age FROM {{.test_schema}}.view_2015_09_13),
    	expected_average_age;
  END IF;
END;
$$ LANGUAGE plpgsql;

SELECT {{.test_schema}}.assert_average_age();
