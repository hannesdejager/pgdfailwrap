# pdgfailwrap [![GoDoc](https://godoc.org/github.com/hannesdejager/pgdfailwrap?status.svg)](https://godoc.org/github.com/hannesdejager/pgdfailwrap)

pdgfailwrap is a wrapper driver for 'postgres' drivers that adds the capability to detect a node that is in readonly mode and reject such a connection in order to be resilient to database failover from a master to slave node. The typical use-case being database maintenance.

It aims in the direction of implementing 2 features found in libpq 10+: [multiple hosts support](https://paquier.xyz/postgresql-2/postgres-10-multi-host-connstr/) and [RW/RO connection selector](https://paquier.xyz/postgresql-2/postgres-10-libpq-read-write/) but currently implements a poor mans version where connections will only be made to read+write nodes i.e. you can't choose. The author needed this
for a production problem and found none of the Go PostgreSQL drivers supported this at the time of writing.

**Only tested with lib/pq at the moment. Use at your own risk.** 

## Usage

You'll import a PostgreSQL driver as per normal:

```go
import (
    "database/sql"

    "github.com/hannesdejager/pgdfailwrap"
    _ "github.com/lib/pq"
)
```

but then do database initialization a bit differently:

```go
func main() {
    var conStrs []string
    // master connection string
    conStrs = append(conStrs, fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", masterHost, masterPort, user, password, dbName))
    // slave connection string
    conStrs = append(conStrs, fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", slaveHost, slavePort, user, password, dbName))
    InitDriver(&pq.Driver{})
    sqlDb, err := sql.Open("postgres-with-failover", strings.Join(conStrs, ","))
    if err != nil {
        log.Fatal(err)
    }
    ...
}
```
