package main

import (
	"bufio"
	"flag"
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
	reciever := flag.String("to", "00000000-0000-0000-0000-000000000000", "recipent of messages")
	// send id to server to check new unread messages
	initMessage := message.Message{"00000000-0000-0000-0000-000000000000", id.String(), nil}
	err := misc.SendMessageTo(conn, &initMessage)
	if err != nil {
		conn.Close()
		log.Fatal(err)
	}

	// TODO: add sync message from server with
	//			 number of unread messages to print
	//       them all
	recvMessage, err := misc.ReadMessageFrom(conn)
	fmt.Printf("%s > %s\n", recvMessage.From, recvMessage.GetMessage())


	// ---
	for {
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
		if len(sendBuf) == 0 {
			continue
		}

	  flag.Parse()
	  recieverId := *reciever

	  sendMessage := message.Message{recieverId, id.String(), sendBuf}
	  err = misc.SendMessageTo(conn, &sendMessage)
	  if err != nil {
	  	conn.Close()
	  	log.Fatal(err)
	  }

	  recvMessage, err = misc.ReadMessageFrom(conn)
	  if err != nil {
	  	conn.Close()
	  	log.Fatal(err)
	  }
	  fmt.Printf("%s> %s\n", recvMessage.From, recvMessage.GetMessage())
	}
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

	handle(conn, id)

}
