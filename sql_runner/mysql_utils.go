package main

import (
	"fmt"
	"strings"
	"regexp"
)

// Formats a single query string primarily for readability
func FormatQueryString(queryString string) string {
	newString := strings.ReplaceAll(queryString, ";\n", ";")
	newString = strings.ReplaceAll(newString, "\t", " ")

	// Remove comment lines
	strSlice := strings.Split(queryString, "\n")
	var newSlice []string
	if len(strSlice) > 0 {
		for _, str := range strSlice {
			//Not a comment, continue
			if strings.Index(str, "--") == -1 {
				newSlice = append(newSlice, str)
			}
		}
	}

	newString = strings.Join(newSlice, " ")

	return strings.Trim(newString, " ")
}

func DelimiterCommandIndex(commandStr string) int {
	return strings.Index(strings.ToLower(strings.Trim(commandStr, " ")), "delimiter")
}

// Whether the command restores the delimiter to a semicolon
// TODO: Terminating line needs to include delimiter and then separate
// "Delimiter ;" to its own line
func IsTerminatingDelimiterLine(commandStr string) (bool, error) {
	terminatorReg, _ := regexp.Compile(`(DELIMITER|delimiter)(\s+);`)
	location := terminatorReg.FindStringIndex(commandStr)
	// delimiterEndsLine := location[1] == len(commandStr)

	if location != nil && location[1] != len(commandStr) {
		return false, fmt.Errorf("IsTerminatingDelimiterLine: delimiter line contains unexpected characters \"%v\"", commandStr)
	}
	return location != nil && location[1] == len(commandStr), nil
}

// Takes a string containing multiple SQL queries and splits each query into its own string
func FormatMultiqueryString(queryString string) []string {
	strSlice := strings.SplitAfter(queryString, ";\n")
	var usableSlice []string

	openDelimiter := false
	var delimitedSlice []string

	for _, str := range strSlice {
		str = FormatQueryString(str)
		if len(str) == 0 {
			continue
		}
		// Handle delimiter changes. All code should be read as one command
		if DelimiterCommandIndex(str) == 0 {
			openDelimiter = true
			delimitedSlice = append(delimitedSlice, str)
		} else if openDelimiter {
			delimitedSlice = append(delimitedSlice, str)
			terminates, _ := IsTerminatingDelimiterLine(str)
			if terminates {
				openDelimiter = false
				usableSlice = append(usableSlice, FormatQueryString(strings.Join(delimitedSlice, " ")))
				// reset slice for next use
				delimitedSlice = nil
			}
		} else {
			usableSlice = append(usableSlice, str)
		}
	}

	return usableSlice
}

// Check if query command is a SELECT by checking the starting SQL keyword.
func IsSelectQuery(commandStr string) bool {
	firstWord := strings.ToLower(strings.Split(commandStr, " ")[0])

	return firstWord == "select"
}