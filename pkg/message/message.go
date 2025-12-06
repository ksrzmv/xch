package message

import (
	"encoding/json"
	"log"
)

type Message struct {
	To 			string
	From 		string
	Msg 		[]byte
}

func Init() *Message {
	var m Message
	m.To 			= ""
	m.From 		= ""
	m.Msg			= make([]byte, 0)

	return &m
}

func (m Message) ToJson() ([]byte, error) {
	log.SetPrefix("message:ToJson:")
	b, err := json.Marshal(m)
	if err != nil {
		log.Println("Error marshalling message to json")
		return nil, err
	}
	return b, nil
}

func FromJson(b []byte) (*Message, error) {
	log.SetPrefix("message:FromJson:")
	m := Init()
	err := json.Unmarshal(b, m)
	if err != nil {
		log.Println("Error unmarshalling message from json")
		return nil, err
	}
	return m, err
}

func (m Message) GetMessage() string {
	return string(m.Msg)
}
