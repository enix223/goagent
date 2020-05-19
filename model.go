package goagent

// Request task request
type Request interface {
	GetID() string
}

// Response task response
type Response struct {
	RequestID string      `json:"request_id"`
	Success   bool        `json:"success"`
	Error     string      `json:"error"`
	Result    interface{} `json:"result"`
}
