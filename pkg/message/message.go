package message

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
