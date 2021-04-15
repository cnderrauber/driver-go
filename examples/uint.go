/*
 * In this test program, we'll create an database and import 1000 records
 * with unsigned integers
 *
 * Authored by <Huo Linhe> linhe.huo@gmail.com
 */
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/taosdata/driver-go/taosSql"
)

const (
	maxLocationSize = 32
	//maxSqlBufSize   = 65480
)

type config struct {
	hostName   string
	serverPort int
	user       string
	password   string
	dbName     string
}

var configPara config
var taosDriverName = "taosSql"
var url string

func init() {
	flag.StringVar(&configPara.hostName, "h", "127.0.0.1", "The host to connect to TDengine server.")
	flag.IntVar(&configPara.serverPort, "p", 6030, "The TCP/IP port number to use for the connection to TDengine server.")
	flag.StringVar(&configPara.user, "u", "root", "The TDengine user name to use when connecting to the server.")
	flag.StringVar(&configPara.password, "P", "taosdata", "The password to use when connecting to the server.")
	flag.StringVar(&configPara.dbName, "d", "taosuint", "Destination database.")
	flag.Parse()
}

func printAllArgs() {
	fmt.Printf("============= args parse result: =============\n")
	fmt.Printf("hostName:             %v\n", configPara.hostName)
	fmt.Printf("serverPort:           %v\n", configPara.serverPort)
	fmt.Printf("usr:                  %v\n", configPara.user)
	fmt.Printf("password:             %v\n", configPara.password)
	fmt.Printf("dbName:               %v\n", configPara.dbName)
	fmt.Printf("================================================\n")
}

func main() {
	printAllArgs()

	url = "root:taosdata@/tcp(" + configPara.hostName + ":" + strconv.Itoa(configPara.serverPort) + ")/"

	test(configPara.dbName)
}

func test(dbName string) {
	db, err := sql.Open(taosDriverName, url)
	if err != nil {
		fmt.Printf("Open database error: %s\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// create an test database with keep option 36500 (100 years to test time before 1970).
	sqlStr := "create database if not exists " + dbName + " keep 36500 days 30"
	fmt.Printf("- %s\n", sqlStr)
	_, err = db.Exec(sqlStr)
	checkErr(err, sqlStr)

	unsignedType(dbName, "uint8", "tinyint unsigned", 0xff)
	unsignedType(dbName, "uint16", "smallint unsigned", 0xffff)
	unsignedType(dbName, "uint32", "int unsigned", 0xffffffff)
	unsignedType(dbName, "uint64", "bigint unsigned", 0xffffffffffffffff)

	sqlStr = "drop database " + dbName
	fmt.Printf("- %s\n", sqlStr)
	_, err = db.Exec(sqlStr)
	checkErr(err, sqlStr)
}

func unsignedType(dbName string, tableName string, typeName string, typeMax uint64) {

	db, err := sql.Open(taosDriverName, url)
	if err != nil {
		fmt.Printf("Open database error: %s\n", err)
		os.Exit(1)
	}
	defer db.Close()

	var sqlStr string

	sqlStr = "create table if not exists " + dbName + "." + tableName + " (ts timestamp, n " + typeName + ")"
	fmt.Printf("- %s\n", sqlStr)
	_, err = db.Exec(sqlStr)
	checkErr(err, sqlStr)

	fmt.Println("\n# Case:", typeName)
	// start time
	const ts1 = "2020-01-02T15:04:05Z"
	st, _ := time.Parse(time.RFC3339, ts1)
	stms := st.UnixNano() / 1000000

	var i uint64 = 0
	var max = ^uint64(0)&typeMax - 1
	for i = 0; i < 10; i++ {
		sqlStr = fmt.Sprintf("insert into %s.%s values(%d,%d)", dbName, tableName, stms+int64(i)*1000, max-i)
		fmt.Printf("- %s\n", sqlStr)
		_, err = db.Exec(sqlStr)
		checkErr(err, sqlStr)
	}

	// select back
	sqlStr = "select last(*) from " + dbName + "." + tableName
	fmt.Printf("- %s\n", sqlStr)

	rows, err := db.Query(sqlStr)
	checkErr(err, sqlStr)

	defer rows.Close()
	fmt.Printf("- %s\n", sqlStr)
	for rows.Next() {
		switch tableName {
		case "uint8":
			var (
				ts string
				n  uint8
			)
			err := rows.Scan(&ts, &n)
			checkErr(err, "rows scan fail")
			fmt.Printf("** last row: (%s, %d)\n", ts, n)
			fmt.Printf("** last n for *%s* is %d\n", typeName, n)
			break
		case "uint16":
			var (
				ts string
				n  uint16
			)
			err := rows.Scan(&ts, &n)
			checkErr(err, "rows scan fail")
			fmt.Printf("** last row: (%s, %d)\n", ts, n)
			fmt.Printf("** last n for *%s* is %d\n", typeName, n)
			break
		case "uint32":
			var (
				ts string
				n  uint32
			)
			err := rows.Scan(&ts, &n)
			checkErr(err, "rows scan fail")
			fmt.Printf("** last row: (%s, %d)\n", ts, n)
			fmt.Printf("** last n for *%s* is %d\n", typeName, n)
			break
		case "uint64":
			var (
				ts string
				n  uint64
			)
			err := rows.Scan(&ts, &n)
			checkErr(err, "rows scan fail")
			fmt.Printf("** last row: (%s, %d)\n", ts, n)
			fmt.Printf("** last n for *%s* is %d\n", typeName, n)
			break
		}
	}
}

func checkErr(err error, prompt string) {
	if err != nil {
		fmt.Printf("ERROR: %s\n", prompt)
		panic(err)
	}
}