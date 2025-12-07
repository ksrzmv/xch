package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

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
	sqlStatement := `CREATE TABLE IF NOT EXISTS users
									  (
											id 			UUID NOT NULL PRIMARY KEY,
											name 		VARCHAR(64),
											passwd 	CHAR(128)
										);`
	sqlStatement += `CREATE TABLE IF NOT EXISTS unread_messages
										(
											id 					BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
											reciever		UUID REFERENCES users(id),
											sender			UUID REFERENCES users(id),
											message 		VARCHAR(255),
											isRead			BOOL NOT NULL DEFAULT FALSE CHECK (isRead = FALSE),
											created_at	TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
										);`
	sqlStatement += `CREATE TABLE IF NOT EXISTS read_messages
										(
											id 				BIGINT PRIMARY KEY,
											reciever	UUID REFERENCES users(id),
											sender		UUID REFERENCES users(id),
											message		VARCHAR(255),
											isRead		BOOL NOT NULL DEFAULT TRUE CHECK (isRead = TRUE),
											read_at		TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
										);`
	_, err := db.Exec(sqlStatement)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("database prepared")
}
