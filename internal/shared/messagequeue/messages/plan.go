package messages

import "eric-cw-hsu.github.io/internal/shared/messagequeue"

type PlanNodeMessage struct {
	Action string                 `json:"action"`
	Index  string                 `json:"index"`
	Key    string                 `json:"key"`
	Data   map[string]interface{} `json:"data"`
}

func (m PlanNodeMessage) Type() string {
	return "plan.node." + m.Action
}

func (m PlanNodeMessage) Encode() ([]byte, error) {
	return messagequeue.EncodeBase(m)
}
