package goagent

// ParseRequestFunc parse request function
type ParseRequestFunc func(payload []byte) (Request, error)

// Broker task broker
type Broker interface {
	// Subscribe subscribe specific task topic, and return the request channel
	// if success, or return error if subscribe failed. The `done` channel is
	// used to notify the broker that the agent will be stopped
	Subscribe(agentID, requestTopic string, done <-chan struct{}, fn ParseRequestFunc) (<-chan Request, error)

	// NotifyResult when the task handler finished processing the request,
	// response will be send back to broker
	NotifyResult(resultTopic string, requestID string, result Response) error
}
