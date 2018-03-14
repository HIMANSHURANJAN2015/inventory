package db

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

var (
	configurations MysqlConfig
	currentDb string
	Db *sql.DB
)	

func Configure(c MysqlConfig, dbName string) {
	configurations = c
	currentDb = dbName
	Connect(dbName)
}

//type NullString sql.NullString

func Connect(dbName string) {
	currentDb = dbName
	var err error
	dsn := dsn(configurations[currentDb].Master, currentDb)
	if Db, err = sql.Open("mysql", dsn); err != nil {
			panic(err)
			//panic("SQL Driver Error", err)
			return
	}
	// Check if is alive
	if err = Db.Ping(); err != nil {
		log.Println("Database Error", err)
	}
}

// DSN returns the Data Source Name
func dsn(ci Config, currentDb string) string {
	// Example: root:@tcp(localhost:3306)/test
	return ci.Username +
		":" +
		ci.Password +
		"@tcp(" +
		ci.Hostname +
		":3306" +
		")/" +
		currentDb
}

type MysqlConfig map[string]ReplicationConfig

type ReplicationConfig struct {
	Master Config `json:"master"`
	Slave Config  `json:"slave"`
}

type Config struct {
	Username  string
	Password  string
	Hostname  string
}


// Inserts and returns the return res of type Result 
func Insert(query string, args ...interface{}) (sql.Result) {
	res, err:= Db.Exec(query, args...)
	log.Println(res)
	log.Println(err)
	if err != nil {
		panic(err)
	}
	return res
}


// errors are deferred until row scan is called
func Row(query string, args ...interface{}) *sql.Row {
	row := Db.QueryRow(query, args...)
	log.Println("query", query)
	log.Println("res", row)
	return row
}

func Select(query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := Db.Query(query, args...)
	return rows, err
}
