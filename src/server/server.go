package main

import (
	"fmt"
	"log"
	"net"

	"database/sql"

	"github.com/ksrzmv/xch/pkg/message"
	"github.com/ksrzmv/xch/pkg/misc"
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

func readMessageFromClient(conn net.Conn) (*message.Message, error) {
	m := message.Init()

	// read client's input
	readBuf := make([]byte, BUFFER_SIZE)
	n, err := conn.Read(readBuf)
	if err != nil {
		return nil, err
	}

	// trims b to size n
	readBuf, err = misc.Trim(readBuf, n)
	if err != nil {
		return nil, err
	}

	// unmarshal input message
	m, err = message.FromJson([]byte(readBuf))
	if err != nil {
		return nil, err
	}
	return m, nil
}

func sendMessageToClient(conn net.Conn, m *message.Message) error {
	sendBuf, err := m.ToJson()
	if err != nil {
		return err
	}
	conn.Write(sendBuf)
	if err != nil {
		return err
	}
	log.Printf("send message '%s' to client", sendBuf)
	return nil
}

// handle(net.Conn, *sql.DB) - handles client connections: reply to requests, stores info in DB
func handle(conn net.Conn) {
	db := dbConnect()
	for {
		m, err := readMessageFromClient(conn)
		if err != nil {
			log.Println(err)
			conn.Close()
			break
		}
		// prints unmarshalled message to stdout
		fmt.Println(*m)

		// insert message into database
		sqlStatement := `
											INSERT INTO messages (sender, reciever, message) VALUES ($1, $2, $3)
											RETURNING id;
										`
		_, err = db.Exec(sqlStatement, m.From, m.To, m.Msg)
		if err != nil {
			log.Println(err)
			conn.Close()
			break
		}

		// sends ack to client
		sendBuf := []byte("ok")
		sendMessage := message.Message{m.From, "00000000-0000-0000-0000-000000000000", sendBuf}
		err = sendMessageToClient(conn, &sendMessage)
		if err != nil {
			log.Println(err)
			conn.Close()
			break
		}
  }
	
}

func dbConnect() *sql.DB {
	pgConninfo := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbName, dbSSLMode)
	// connect to db
	db, err := sql.Open(dbDriver, pgConninfo)
	if err != nil {
		db.Close()
		log.Fatal(err)
	}

	// pings db to ensure that's we've successfully connected
	err = db.Ping()
	if err != nil {
		db.Close()
		log.Fatal(err)
	}
	log.Println("db successfully connected")

	return db
}

// create messages table
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

// listen on socket
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
