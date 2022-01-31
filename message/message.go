package message

import (
	"TP2_Minimum_Spanning_Tree/nodeState"
	"encoding/json"
	"fmt"
	"io"
)

type Type string

const (
	Connect    Type = "Connect"
	Initiate   Type = "Initiate"
	Test       Type = "Test"
	Accept     Type = "Accept"
	Reject     Type = "Reject"
	Report     Type = "Report"
	ChangeRoot Type = "ChangeRoot"
)

type Message struct {
	Type       Type
	FragmentId string
	Level      int
	Weight     int
	State      nodeState.NodeState
	Port       string
}

func New(messageType Type, fragmentId string, level int, weight int, state nodeState.NodeState, port string) Message {
	return Message{
		Type:       messageType,
		FragmentId: fragmentId,
		Level:      level,
		Weight:     weight,
		State:      state,
		Port:       port,
	}
}

func (message Message) ToJSON() ([]byte, error) {
	return json.Marshal(message)
}

func FromJSON(r io.Reader) (Message, error) {
	decoder := json.NewDecoder(r)
	var message Message
	var err = decoder.Decode(&message)
	return message, err
}

func (message Message) ToString() string {
	return fmt.Sprintf("Message(%+v)", message)
}
