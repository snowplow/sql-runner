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
)

const msNetwork = "tcp"

// Struct of client and target databases
// Target is an anonymous Target struct defined in playbook.go
type MySQLTarget struct {
	Target
	Client *sql.DB
	clientConfig *mysql.Config
}

// Checks whether target database can connect
func (mt MySQLTarget) IsConnectable() bool {
	client := mt.Client
	var result int
	err := client.QueryRow("SELECT 1").Scan(&result) // test connection
	if err != nil && VerbosityOption > 0 {
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

func (mt MySQLTarget) GetTarget() Target {
	return mt.Target
}

// Run a query against the target
func (mt MySQLTarget) RunQuery(query ReadyQuery, dryRun bool, showQueryOutput bool) QueryStatus {
	var err error = nil
	if dryRun {
		address := mt.clientConfig.Addr
		if mt.IsConnectable() {
			if VerbosityOption == MAX_VERBOSITY {
				log.Printf("SUCCESS: Able to connect to target database, %s\n.", address)
			}
		} else if VerbosityOption > 0 {
			log.Printf("ERROR: Cannot connect to target database, %s\n.", address)
		}
		return QueryStatus{query, query.Path, 0, nil}
	}

	affected := 0

	// sql.Exec cannot handle commands formatted over multiple lines
	queryStrings := FormatMultiqueryString(query.Script)

	txn, err := mt.Client.Begin()
	if err == nil {
		defer func() {
			// If transaction was already committed, this will do nothing
			_ = txn.Rollback()
		}()

		for _, q := range queryStrings {
			// This should ensure that we return an accurate Rows Affected value for SELECT queries to match postgres_target's Rows Affected policy
			if IsSelectQuery(q) {
				var rows *sql.Rows
				rows, err = mt.Client.Query(q)
				if err == nil {
					affected, _ = interpretRows(rows, showQueryOutput)
				}
			} else {
				var res sql.Result
				res, err = mt.Client.Exec(q)
				if err == nil {
					affectedInt64, _ := res.RowsAffected()
					affected = int(affectedInt64)
				}
			}
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
func interpretRows(rows *sql.Rows, shouldPrintTable bool) (affected int, funcErr error) {
	defer rows.Close()

	if !shouldPrintTable {
		for rows.Next() {
			affected += 1
		}
	} else {
		affected, funcErr = printMSTable(rows)
	}
	return
}

// Print table produced by sql.DB.Query
func printMSTable(rows *sql.Rows) (affected int, funcErr error) {
	columns, colErr := rows.Columns()
	if colErr != nil {
		return affected, errors.New("Unable to read columns")
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
			log.Printf("ERROR | printMSTable: %s", err)
		} else {
			affected += 1
			table.Append(stringList(strs))
		}
	}

	table.Render()
	return
}