package main

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"

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
	slog.Error("Error processing request", slog.String("remote_addr", conn.RemoteAddr().String()), slog.Any("err_msg", err))
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

		// TODO: move this transaction to PGSQL fuction
		tx, err := db.Begin()
		if err != nil {
			errorHandler(conn, db, err)
			return err
		}
		defer tx.Rollback()

		sqlStatement = 	`
											INSERT INTO read_messages(id, sender, reciever, message) VALUES ($1, $2, $3, $4);
										`
		_, err = tx.Exec(sqlStatement, messageId, sender, reciever, msg)
		if err != nil {
			slog.Error("failed insert message to read_messages", slog.Any("err_msg", err))
			return err
		}
		sqlStatement =  `
											DELETE FROM unread_messages WHERE id = $1;
										`
		_, err = tx.Exec(sqlStatement, messageId)
		if err != nil {
			slog.Error("failed delete message to unread_messages", slog.Any("err_msg", err))
			return err
		}

		if err = tx.Commit(); err != nil {
			errorHandler(conn, db, err)
			return err
		}
		// --------
	}

	err = unreadMessages.Err()
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
	slog.Info("recieved connection", slog.String("remote_addr", conn.RemoteAddr().String()))
	db := dbConnect()


	// check for unread messages
	initMessage, err := misc.ReadMessageFrom(conn)
	if err != nil {
		errorHandler(conn, db, err)
		return
	}
	slog.Info("recieved init message", slog.String("user_id", initMessage.From), slog.String("remote_addr", conn.RemoteAddr().String()))

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
			slog.Debug("added user to DB", slog.String("user_id", id))
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
			errorHandler(conn, db, err)
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
			errorHandler(conn, db, err)
			break
		}

		// sends ack to client
		sendBuf := []byte("ok")
		sendMessage := message.Message{m.From, "00000000-0000-0000-0000-000000000000", sendBuf}
		err = misc.SendMessageTo(conn, &sendMessage)
		if err != nil {
			errorHandler(conn, db, err)
			break
		}
		slog.Info("sent message to user", slog.String("message", sendMessage.GetMessage()), slog.String("user_id", m.From))
  }
	
}

// listen on socket
func listen() net.Listener {
	socket := fmt.Sprintf("%s:%s", listenHost, listenPort)
	ln, err := net.Listen(listenProto, socket)
	if err != nil {
		panic(err)
	}
	slog.Info("listening...", slog.String("proto", listenProto), slog.String("socket", socket))

	return ln
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	log.SetPrefix("main: ")
	db := dbConnect()
	dbPrepare(db)

	ln := listen()

	for {
		conn, err := ln.Accept()
		if err != nil {
			errorHandler(conn, db, err)
			continue
		}

		go handle(conn)
	}
}
