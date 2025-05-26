package messagequeue

import "encoding/json"

type Message interface {
	Type() string
	Encode() ([]byte, error)
}

type BaseMessage struct {
	Type string          `json:"type"`
	Body json.RawMessage `json:"body"`
}

func EncodeBase(msg Message) ([]byte, error) {
	body, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	wrapper := BaseMessage{
		Type: msg.Type(),
		Body: body,
	}

	return json.Marshal(wrapper)
}
