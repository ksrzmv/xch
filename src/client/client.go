package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/google/uuid"

	"github.com/ksrzmv/xch/pkg/message"
	"github.com/ksrzmv/xch/pkg/misc"
)

const (
	xchHost           = "localhost"
	xchPort           = "3333"
	idFilename        = ".xch-id"
	uuidLengthSymbols = 36
	
	INPUT_BUFFER			= 128
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

func handle(conn net.Conn, id uuid.UUID) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("> ")

	// TODO: possible overflow. limit the input text string.
	// TODO: implement the multiline messages
	readBuf, err := reader.ReadString('\n')
	if err != nil {
		conn.Close()
		log.Fatal(err)
	}
	
	// trims trailing '\n'
	sendBuf := []byte(readBuf)[:len(readBuf)-1]
	m := message.Message{"00000000-0000-0000-0000-000000000000", id.String(), sendBuf}
	sendBuf, err = m.ToJson()
  if err != nil {
  	conn.Close()
  	log.Fatal(err)
  }

  conn.Write(sendBuf)

	answerBuf := make([]byte, 1024)
	readBytes, err := conn.Read(answerBuf)
	if err != nil {
		conn.Close()
		log.Fatal(err)
	}

	answerBuf, err = misc.Trim(answerBuf, readBytes)
	if err != nil {
		conn.Close()
		log.Fatal()
	}

	recvMessage, err := message.FromJson(answerBuf)
	if err != nil {
		conn.Close()
		log.Fatal(err)
	}
	fmt.Println(recvMessage.GetMessage())
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

	for {
		handle(conn, id)
	}

}
