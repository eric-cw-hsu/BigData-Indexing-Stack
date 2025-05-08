package types

type PlanMessage struct {
	Action string                 `json:"action"` // "create" | "update" | "delete"
	Key    string                 `json:"key"`
	Data   map[string]interface{} `json:"data,omitempty"`
}
