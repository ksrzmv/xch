package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"

	"github.com/google/uuid"
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

func errorHandler(conn net.Conn, db *sql.DB, err error) {
	db.Close()
	conn.Close()
	log.Println(err)
}

func handleUnreadMessages(conn net.Conn, db *sql.DB, id uuid.UUID) error {
	sqlStatement := `
										SELECT id, sender, reciever, message FROM unread_messages WHERE reciever = $1 AND
										  isRead = false;
									`
	unreadMessages, err := db.Query(sqlStatement, id)
	if err != nil {
		errorHandler(conn, db, err)
		return err
	}

	haveUnreadMessages := false

	defer unreadMessages.Close()
	for unreadMessages.Next() {
		var messageId	string
		var sender 		string
		var reciever 	string
		var msg				[]byte
		err = unreadMessages.Scan(&messageId, &sender, &reciever, &msg)
		if err != nil {
			return err
		}

		sendMessage := message.Message{sender, reciever, msg}
		err = misc.SendMessageTo(conn, &sendMessage)
		if err != nil {
			errorHandler(conn, db, err)
			return err
		}
		haveUnreadMessages = true

		// TODO: make transaction
		// TODO: move this transaction to PGSQL fuction
		sqlStatement = 	`
											INSERT INTO read_messages(id, sender, reciever, message) VALUES ($1, $2, $3, $4);
										`
		_, err = db.Query(sqlStatement, messageId, sender, reciever, msg)
		if err != nil {
			return err
		}
		sqlStatement = 	`
											DELETE FROM unread_messages WHERE id = $1;
										`
		_, err = db.Query(sqlStatement, messageId)
		if err != nil {
			return err
		}
		// --------
	}

	err = unreadMessages.Err()
	log.Println(err)
	if err != nil {
		errorHandler(conn, db, err)
		return err
	}

	if haveUnreadMessages == false {
		return sql.ErrNoRows
	}
	return nil
}

// handle(net.Conn) - handles client connections: reply to requests, stores info in DB
func handle(conn net.Conn) {
	log.SetPrefix("handle conn: ")
	log.Printf("recieved connection from %s\n", conn.RemoteAddr().String())
	db := dbConnect()


	// check for unread messages
	initMessage, err := misc.ReadMessageFrom(conn)
	if err != nil {
		errorHandler(conn, db, err)
		return
	}
	log.Println("recieved init message")

	id := initMessage.From
	sqlStatement := `
										SELECT id FROM users WHERE id = $1;
									`
	tmpBuf := make([]byte, 32)
	err = db.QueryRow(sqlStatement, id).Scan(&tmpBuf)
	switch err {
		case sql.ErrNoRows:
			sqlStatement = 	`
												INSERT INTO users(id) VALUES ($1);
											`
			_, err = db.Exec(sqlStatement, id)
			if err != nil {
				errorHandler(conn, db, err)
			}
	  case nil:
			tmpId, err := uuid.Parse(id)
			if err != nil {
				errorHandler(conn, db, err)
			}
			err = handleUnreadMessages(conn, db, tmpId)
			if err != nil {
				err = misc.SendMessageTo(conn, &message.Message{initMessage.From, initMessage.To, []byte("No unread messages")})
				if err != nil {
					errorHandler(conn, db, err)
				}
			}
		default:
			errorHandler(conn, db, err)
	}

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
											INSERT INTO unread_messages (sender, reciever, message) VALUES ($1, $2, $3)
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
