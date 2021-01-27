-- Test file: assert.sql

DELIMITER $$
DROP FUNCTION IF EXISTS {{.test_schema}}.assert_average_age;
CREATE FUNCTION {{.test_schema}}.assert_average_age(age INT) RETURNS INT
BEGIN
  DECLARE expected_average_age INT DEFAULT 23;
  IF age <> expected_average_age THEN
    SIGNAL SQLSTATE 'HY000' SET MESSAGE_TEXT = 'average_age does not match expected age';
  END IF;
  RETURN (expected_average_age);
END $$

DELIMITER ;

SELECT {{.test_schema}}.assert_average_age(a.average_age)
FROM (
  SELECT average_age FROM {{.test_schema}}.view_{{.test_date}}
) as a;
