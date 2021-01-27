-- Test file: 1.sql
DROP DATABASE IF EXISTS {{.test_schema}};

CREATE DATABASE {{.test_schema}};

USE {{.test_schema}};

CREATE TABLE table1 (
    age INT,
    firstName VARCHAR(255),
    city VARCHAR(255),
    country VARCHAR(255)
);