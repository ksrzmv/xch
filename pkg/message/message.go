package message

import (
	"encoding/json"
	"log"
	"log/slog"
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
	log.SetPrefix("message:ToJson: ")
	b, err := json.Marshal(m)
	if err != nil {
		slog.Error("Error marshalling message to json", slog.Any("err_msg", err))
		return nil, err
	}
	return b, nil
}

func FromJson(b []byte) (*Message, error) {
	log.SetPrefix("message:FromJson: ")
	m := Init()
	err := json.Unmarshal(b, m)
	if err != nil {
		slog.Error("Error unmarshalling message from json", slog.Any("err_msg", err))
		return nil, err
	}
	return m, err
}

func (m Message) GetMessageRaw() []byte {
	return m.Msg
}

func (m Message) GetMessage() string {
	return string(m.Msg)
}

func (m Message) GetTo() string {
	return m.To
}

func (m Message) GetFrom() string {
	return m.From
}

