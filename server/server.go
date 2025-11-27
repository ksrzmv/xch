package main

import (
	"net"
	"log"
	"fmt"
	"errors"

	"database/sql"
	_ "github.com/lib/pq"
)

const (
	dbDriver 		= "postgres"
	dbHost 			= "localhost"
	dbPort 			= "5432"
	dbUser			= "postgres"
	dbName 			= "xch"
	dbSSLMode 	= "disable"
	BUFFER_SIZE = 1024
)


// trim([]byte, int) ([]byte, error) - trims input byte slice to new byte slice of size n
func trim(b []byte, n int) ([]byte, error) {
	if (n > len(b)) {
		return nil, errors.New("trimmed slice should have size smaller of or equal to origin slice")
	}
	result := make([]byte, n, n)
	for i := 0; i < n; i++ {
		result[i] = b[i]
	}
	return result, nil
}

// handle(net.Conn, *sql.DB) - handles client connections: reply to requests, stores info in DB
func handle(conn net.Conn, db *sql.DB) {
	for {
		b := make([]byte, BUFFER_SIZE)
		n, err := conn.Read(b)
		if err != nil {
			log.Print(err)
			conn.Close()
			break
		}

		// trims b to size n
		b, err = trim(b, n)
		if err != nil {
			log.Println(err)
			continue
		}

		sqlStatement := fmt.Sprintf("INSERT INTO messages (message, ip_from) VALUES ('%s', '%s');", string(b), conn.RemoteAddr().String())

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
	sqlStatement := `CREATE TABLE IF NOT EXISTS messages 
										(
											id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
											message VARCHAR(255),
											ip_from VARCHAR(255),
											time		TIMESTAMPTZ NOT NULL DEFAULT NOW()
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
