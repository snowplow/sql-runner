-- Test file: assert.sql

CREATE OR REPLACE PROCEDURAL LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION {{.test_schema}}.assert_average_age() RETURNS VOID AS $$
DECLARE
  expected_average_age CONSTANT integer := 23;
BEGIN
  {{/* Update view names to today's date or Travis will error */}}
  IF (SELECT average_age <> expected_average_age FROM {{.test_schema}}.view_{{.test_date}}) THEN
    RAISE EXCEPTION 'Average_age % does not match expected age %',
    	(SELECT average_age FROM {{.test_schema}}.view_{{.test_date}}),
    	expected_average_age;
  END IF;
END;
$$ LANGUAGE plpgsql;

SELECT {{.test_schema}}.assert_average_age();
