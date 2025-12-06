package main

import (
	"fmt"
	"log"
	"net"

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

	listenHost  = "localhost"
	listenPort  = "3333"
	listenProto = "tcp"
)

// handle(net.Conn) - handles client connections: reply to requests, stores info in DB
func handle(conn net.Conn) {
	log.SetPrefix("handle conn: ")
	log.Printf("recieved connection from %s\n", conn.RemoteAddr().String())
	db := dbConnect()
	for {
		m, err := misc.ReadMessageFrom(conn)
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
		err = misc.SendMessageTo(conn, &sendMessage)
		if err != nil {
			log.Println(err)
			conn.Close()
			break
		}
		log.Printf("sent message `%s` to `%s`\n", sendMessage.GetMessage(), m.From)
  }
	
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
