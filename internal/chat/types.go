package chat

type InboundMsg struct {
	Payload string `json:"payload"`
	User    string `json:"user,omitempty"`
}

type OutboundMsg struct {
	ID        string `json:"id"`
	Payload   string `json:"payload"`
	User      string `json:"user"`
	Timestamp int64  `json:"timestamp"`
}
