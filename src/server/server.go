package main

import (
	"errors"
	"fmt"
	"log"
	"net"

	"database/sql"

	"github.com/ksrzmv/xch/pkg/message"
	_ "github.com/lib/pq"
)

const (
	dbDriver    = "postgres"
	dbHost      = "localhost"
	dbPort      = "5432"
	dbUser      = "postgres"
	dbName      = "xch"
	dbSSLMode   = "disable"

	BUFFER_SIZE = 1024

	listenHost  = "localhost"
	listenPort  = "3333"
	listenProto = "tcp"
)

// trim([]byte, int) ([]byte, error) - trims input byte slice to new byte slice of size n
func trim(b []byte, n int) ([]byte, error) {
	if n > len(b) {
		return nil, errors.New("trimmed slice should have size smaller of or equal to origin slice")
	}
	result := make([]byte, n, n)
	for i := 0; i < n; i++ {
		result[i] = b[i]
	}
	return result, nil
}

// handle(net.Conn, *sql.DB) - handles client connections: reply to requests, stores info in DB
func handle(conn net.Conn) {
	db := dbConnect()
	for {
		m := message.Init()
		readBuf := make([]byte, BUFFER_SIZE)

		n, err := conn.Read(readBuf)
		if err != nil {
			log.Println(err)
			conn.Close()
			break
		}

		// trims b to size n
		readBuf, err = trim(readBuf, n)
		if err != nil {
			log.Println(err)
			continue
		}

		m, err = message.FromJson([]byte(readBuf))
		if err != nil {
			log.Println(err)
			conn.Close()
			break
		}
		fmt.Println(*m)

		// insert message into database
		sqlStatement := `
					INSERT INTO messages (sender, reciever, message) VALUES ($1, $2, $3)
					RETURNING id`

		_, err = db.Exec(sqlStatement, m.From, m.To, m.Msg)
		if err != nil {
			log.Println(err)
			conn.Close()
			break
		}

		sendBuf := []byte("ok")
		conn.Write(sendBuf)
		if err != nil {
			log.Println(err)
			conn.Close()
			break
		}
		log.Printf("send message '%s' to client", sendBuf)

  }
	
}

func dbConnect() *sql.DB {
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

	return db
}

func dbPrepare(db *sql.DB) {
	sqlStatement := `CREATE TABLE IF NOT EXISTS messages 
										(
											id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
											sender		UUID NOT NULL,
											reciever	UUID NOT NULL,
											message 	VARCHAR(255),
											isRead		BOOL NOT NULL DEFAULT FALSE,
											time			TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
										);`

	_, err := db.Exec(sqlStatement)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("database prepared")
}

func listen() net.Listener {
	socket := fmt.Sprintf("%s:%s", listenHost, listenPort)
	ln, err := net.Listen(listenProto, socket)
	if err != nil {
		panic(err)
	}
	log.Printf("start listening on %s %s\n", listenProto, socket)

	return ln
}

func main() {
	db := dbConnect()
	dbPrepare(db)

	ln := listen()

	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}

		go handle(conn)
	}
}
