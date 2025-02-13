# Go Connector for TDengine

[![Build Status](https://cloud.drone.io/api/badges/taosdata/driver-go/status.svg)](https://cloud.drone.io/taosdata/driver-go)

English | [简体中文](README-CN.md)

[TDengine] provides Go `database/sql` driver as [`taosSql`][driver-go].

## Install

Go 1.14+ is highly recommended for newly created projects.

```sh
go mod init taos-demo
```

import taosSql：

```go
import (
    "database/sql"
    _ "github.com/taosdata/driver-go/taosSql"
)
```

Use `go mod` for module management:

```sh
go mod tidy
```

Or `go get` to directly install it:

```sh
go get github.com/taosdata/driver-go/taosSql
```

Use `win` branch if you are using the module in Windows os, since STMT-series APIs are not fully supported currently.

```sh
go get github.com/taosdata/driver-go/taosSql@win
```

## Usage

### `database/sql` Standard

A simple use case：

```go
import (
	"fmt"
	"database/sql"
	_ "github.com/taosdata/driver-go/taosSql"
)
func main() {
	var taosuri = "root:taosdata/tcp(localhost:6030)/"
	taos, err := sql.Open("taosSql", taosuri)
	if err != nil {
		fmt.Println("failed to connect TDengine, err:", err)
		return
	}
	defer taos.Close()
	taos.Exec("create database if not exists test")
	taos.Exec("use test")
	taos.Exec("create table if not exists tb1 (ts timestamp, a int)")
	_, err = taos.Exec("insert into tb1 values(now, 0)(now+1s,1)(now+2s,2)(now+3s,3)")
	if err != nil {
		fmt.Println("failed to insert, err:", err)
		return
	}
	rows, err := taos.Query("select * from tb1")
	if err != nil {
		fmt.Println("failed to select from table, err:", err)
		return
	}

	defer rows.Close()
	for rows.Next() {
		var r struct {
			ts time.Time
			a  int
		}
		err := rows.Scan(&r.ts, &r.a)
		if err != nil {
			fmt.Println("scan error:\n", err)
			return
		}
		fmt.Println(r.ts, r.a)
	}
}
```

APIs that are worthy to have a check:

- `sql.Open(DRIVER_NAME string, dataSourceName string) *DB`

  This API will create a `database/sql` DB object, results with type `*DB`. `DRIVER_NAME` should be setted as `taosSql`, and `dataSourceName` should be an URI like `user:password@/tcp(host:port)/dbname`. For HA use case, use `user:password@/cfg/dbname` to apply configs in `/etc/taos/taos.cfg`。

- `func (db *DB) Exec(query string, args ...interface{}) (Result, error)`

  Execute non resultset SQLs, like `create`, `alter` etc.

- `func (db *DB) Query(query string, args ...interface{}) (*Rows, error)`

  Execute an query with resultset.

- `func (db *DB) Close() error`

  Close an DB object and disconnect.

### Subscription

Open DB:

```go
Open(dbname string) (db DB, err error)
```

Subscribe:

```go
type DB interface {
	Subscribe(restart bool, name string, sql string, interval time.Duration) (Topic, error)
	Close() error
}
```

Topic:

```go
type Topic interface {
	Consume() (driver.Rows, error)
	Unsubscribe(keepProgress bool)
}
```

Check sample code for subscription at [`examples/taoslogtail.go`](https://github.com/taosdata/driver-go/blob/master/examples/taoslogtail/taoslogtail.go)。

[driver-go]: https://github.com/taosdata/driver-go
[TDengine]: https://github.com/taosdata/TDengine