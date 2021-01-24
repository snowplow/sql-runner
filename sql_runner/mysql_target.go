package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/olekukonko/tablewriter"
	"log"
	"net"
	"os"
	"strings"
)

const msNetwork = "tcp"

// Struct of client and target databases
// Target is an anonymous Target struct defined in playbook.go
type MySQLTarget struct {
	Target
	Client *sql.DB
	clientConfig *mysql.Config
}

func (mt MySQLTarget) IsConnectable() bool {
	client := mt.Client
	var result int
	err := client.QueryRow("SELECT 1").Scan(&result) // test connection
	if err != nil {
		fmt.Println("ERROR:", err)
	}

	return err == nil && result == 1
}

func NewMySQLTarget(target Target) *MySQLTarget {
	targetConfig := mysql.NewConfig()
	if len(target.Username) > 0 {
		targetConfig.User = target.Username
		targetConfig.Passwd = target.Password
	}
	targetConfig.Net = msNetwork
	targetConfig.Addr = target.Host + ":" + target.Port
	targetConfig.DBName = target.Database

	targetConfig.Timeout = SQL_dialTimeout
	targetConfig.ReadTimeout = SQL_readTimeout

	// TLS configuration
	var tlsConfig *tls.Config
	if target.Ssl == true {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
		}

		// Registers custom TLS configuration under name "customTLS" to be used in 
		// MySQL config
		err := mysql.RegisterTLSConfig("customTLS", tlsConfig)
		if err != nil {
			log.Fatal(err)
		}

		targetConfig.TLSConfig = "customTLS"
	}
	
	// Dialer configuration
	mysql.RegisterDialContext(msNetwork,
		func(ctx context.Context, addr string) (net.Conn, error) {
			cn, err := net.DialTimeout(msNetwork, addr, SQL_dialTimeout)
			if err != nil {
				return nil, err
			}
			return cn, cn.(*net.TCPConn).SetKeepAlive(true)
		})
		
	db, err := sql.Open("mysql", targetConfig.FormatDSN())

	if err != nil {
		log.Fatal(err)
	}

	return &MySQLTarget{target, db, targetConfig}
}

// Getter required by run.runSteps for all database targets
func (mt MySQLTarget) GetTarget() Target {
	return mt.Target
}

// Check if query is a SELECT by checking the starting SQL command.
// This should ensure that we return an accurate Rows Affected value for SELECT queries to match postgres_target's Rows Affected policy
func (query ReadyQuery) isSelectQuery() bool {
	firstWord := strings.ToLower(strings.Split(query.Script, " ")[0])

	return firstWord == "select"
}

// Run a query against the target
func (mt MySQLTarget) RunQuery(query ReadyQuery, dryRun bool, showQueryOutput bool) QueryStatus {
	var err error = nil
	if dryRun {
		address := mt.clientConfig.Addr
		if mt.IsConnectable() {
			log.Printf("SUCCESS: Able to connect to target database, %s\n.", address)
		} else {
			log.Printf("ERROR: Cannot connect to target database, %s\n.", address)
		}
		return QueryStatus{query, query.Path, 0, nil}
	}

	var affected int64 = 0

	if query.isSelectQuery() {
		rows, err := mt.Client.Query(query.Script)
		if err == nil {
			affected, _ = interpretRows(rows, showQueryOutput)
		}
	} else {
		res, err := mt.Client.Exec(query.Script)
		if err == nil {
			affected, _ = res.RowsAffected()
		}
	}

	return QueryStatus{query, query.Path, affected, err}
}

// Convert slice of string pointers to slice of strings to be used with tablewriter
func stringList(spList []*string) []string {
	sList := make([]string, len(spList))
	for i := range spList {
		if spList[i] != nil {
			sList[i] = *spList[i]
		} else {
			sList[i] = "NULL"
		}
	}
	return sList
}

// Interpret sql.Rows results to ensure all rows are counted
func interpretRows(rows *sql.Rows, shouldPrintTable bool) (affected int64, funcErr error) {
	defer rows.Close()

	if !shouldPrintTable {
		for rows.Next() {
			affected += 1
		}
	} else {
		affected, funcErr = printTable(rows)
	}
	return
}

// Print table produced by sql.DB.Query
func printTable(rows *sql.Rows) (affected int64, funcErr error) {
	columns, colErr := rows.Columns()
	if colErr != nil {
		return affected, errors.New("Unable to read columns")
		log.Printf("ERROR | printTable: %s", colErr)
	}

	log.Printf("QUERY OUTPUT:\n")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(columns)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")

	strs := make([]*string, len(columns))
	vals := make([]interface{}, len(columns))
	for rows.Next() {
		for i := range vals {
			vals[i] = &strs[i]
		}
		if err := rows.Scan(vals...); err != nil {
			log.Printf("ERROR | printTable: %s", err)
		} else {
			affected += 1
			table.Append(stringList(strs))
		}
	}

	table.Render()
	return
}