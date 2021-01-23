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
	if showQueryOutput {
		var rows *sql.Rows
		rows, err := mt.Client.Query(query.Script)
		if err != nil {
			log.Printf("ERROR: %s.", err)
			return QueryStatus{query, query.Path, affected, err}
		}
		// TODO: Implement printTable
		// affected, _ = printTable(rows)
	} else {
		res, err := mt.Client.Exec(query.Script)
		if err == nil {
			affected, _ = res.RowsAffected()
		}
	}

	return QueryStatus{query, query.Path, affected, err}
}