package main

import (
	"net"
	"log"
	"fmt"

	"database/sql"
	_ "github.com/lib/pq"
)

const (
	dbDriver 	= "postgres"
	dbHost 		= "localhost"
	dbPort 		= "5432"
	dbUser		= "postgres"
	dbName 		= "xch"
	dbSSLMode = "disable"
)


func tr(b []byte, n int) []byte {
	result := make([]byte, n, n)
	for i := 0; i < n; i++ {
		result[i] = b[i]
	}
	return result
}

func handle(conn net.Conn, db *sql.DB) {
	for {
		b := make([]byte, 32)
		n, err := conn.Read(b)
		if err != nil {
			log.Print(err)
			conn.Close()
			break
		}

		b = tr(b, n)

		sqlStatement := fmt.Sprintf("INSERT INTO users (message, ip_from) VALUES ('%s', '%s');", string(b), conn.RemoteAddr().String())

		fmt.Println(sqlStatement)

		_, err = db.Exec(sqlStatement)
		if err != nil {
			log.Println(err)
		}

		conn.Write(b)
		if err != nil {
			log.Print(err)
			conn.Close()
			break
		}

		log.Printf("sent message `%s` to %s\n", b, conn.RemoteAddr().String())
	}
}

func dbPrepare(db *sql.DB) {
	sqlStatement := `CREATE TABLE IF NOT EXISTS users 
										(
											id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
											message VARCHAR(255),
											ip_from VARCHAR(255)
										);`

	_, err := db.Exec(sqlStatement)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("database prepared")
}

func main() {
	pgConninfo := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s",
														dbHost, dbPort, dbUser, dbName, dbSSLMode)
	db, err := sql.Open(dbDriver, pgConninfo)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("db successfully connected")

	dbPrepare(db)

	ln, err := net.Listen("tcp", "localhost:3333")
	if err != nil {
		panic(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}

		go handle(conn, db)
	}
}
