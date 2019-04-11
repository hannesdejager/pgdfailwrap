// Package pgdfailwrap is a wrapper driver for 'postgres' drivers that adds the capability to detect a node that is in readonly mode
// and reject such a connection in order to be resilient to database failover from master to slave. The typical use
// case here is database maintenance.
//
// Example use:
//
//	import (
//		"database/sql"
//
//		_ "github.com/lib/pq"
//	)
//
//	func main() {
//		var conStrs []string
//		conStrs = append(conStrs, fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", masterHost, masterPort, user, password, dbName))
//		conStrs = append(conStrs, fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", slaveHost, slavePort, user, password, dbName))
//		InitDriver(&pq.Driver{})
//		sqlDb, err := sql.Open("postgres-with-failover", strings.Join(conStrs, ","))
//		if err != nil {
//			log.Fatal(err)
//		}
//		...
//	}
//
package pgdfailwrap

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
)

// Driver satisfies the sql/driver/Driver interface
type Driver struct {
	delegate driver.Driver
}

// InitDriver initializes the wrapper driver taking an already initialized postgres driver as argument.
func InitDriver(delegate driver.Driver) {
	sql.Register("postgres-with-failover", &Driver{
		delegate: delegate,
	})
}

// Open will be called by sql.Open. The connection string is comma seperated pairs of connection strings
// that is compatible with the delegate driver being used i.e. conStr[,conStr...]
func (d *Driver) Open(connStr string) (driver.Conn, error) {
	c := strings.Split(connStr, ",")
	var err error
	var con driver.Conn
	for _, s := range c {
		con, err = d.tryOpen(s)
		if err == nil {
			return con, nil
		}
	}
	return nil, err
}

func (d *Driver) tryOpen(pqConnStr string) (driver.Conn, error) {

	db, err := sql.Open("postgres", pqConnStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SHOW transaction_read_only")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, rows.Err()
	}

	var readonlyIndicator string
	err = rows.Scan(&readonlyIndicator)
	if err != nil {
		return nil, err
	}
	if readonlyIndicator == "on" {
		return nil, fmt.Errorf("database in readonly mode")
	}
	db.Close()

	return d.delegate.Open(pqConnStr)
}
