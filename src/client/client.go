package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/google/uuid"

	"github.com/ksrzmv/xch/pkg/message"
)

const (
	xchHost           = "localhost"
	xchPort           = "3333"
	idFilename        = ".xch-id"
	uuidLengthSymbols = 36
)

func getIdFromFile(filepath string) uuid.UUID {
	var id uuid.UUID
	fd, err := os.OpenFile(filepath, os.O_RDONLY, 0600)
	if err != nil {
		fd, err = os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			fd.Close()
			log.Fatal(err)
		}

		id = uuid.New()
		fmt.Fprintf(fd, "%s", id)
	} else {
		buf := make([]byte, uuidLengthSymbols)
		_, err = fd.Read(buf)
		if err != nil {
			fd.Close()
			log.Fatal(err)
		}
		err = id.UnmarshalText(buf)
		if err != nil {
			fd.Close()
			log.Fatal(err)
		}
	}
	fd.Close()
	return id
}

func main() {
	homePath, isPresent := os.LookupEnv("HOME")
	if isPresent == false {
		log.Println("HOME env is not set. Current directory will be treat as home dir")
		homePath = "."
	}

	idFilepath := fmt.Sprintf("%s/%s", homePath, idFilename)

	id := getIdFromFile(idFilepath)

	fmt.Println(id)

	socket := fmt.Sprintf("%s:%s", xchHost, xchPort)
	conn, err := net.Dial("tcp", socket)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("successfully connected to %s\n", socket)

	m, err := json.Marshal(message.Message{"79e60d11-f430-4e7c-9202-dbdb58239131", id.String(), []byte("hello")})
	if err != nil {
		conn.Close()
		log.Fatal(err)
	}

	conn.Write(m)

	conn.Close()
}
