package goagent

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// MessageHandleFunc Task handle function
type MessageHandleFunc func(ctx context.Context, request Request) <-chan Response

type defaultRequest struct {
}

func (defaultRequest) GetID() string {
	return uuid.New().String()
}

// Options agent config
type Options struct {
	ID               string
	Broker           Broker
	TaskRequestTopic string
	TaskResultTopic  string
	Parallel         int
	HandleFunc       MessageHandleFunc
	ParseRequestFunc ParseRequestFunc
	TaskTimeout      time.Duration
	Logger           Logger
}

// Option config option
type Option func(o *Options)

// SetBroker set broker
func SetBroker(b Broker) Option {
	return func(o *Options) {
		o.Broker = b
	}
}

// SetAgentID set agent id
func SetAgentID(id string) Option {
	return func(o *Options) {
		o.ID = id
	}
}

// SetTaskRequestTopic set task request topic
func SetTaskRequestTopic(topic string) Option {
	return func(o *Options) {
		o.TaskRequestTopic = topic
	}
}

// SetTaskResultTopic set task result topic
func SetTaskResultTopic(topic string) Option {
	return func(o *Options) {
		o.TaskResultTopic = topic
	}
}

// SetParallel set the number of parallel handlers of the agent
// If current task handlers reach the limit, all incoming message will be block
// until any task handler finished
func SetParallel(n int) Option {
	return func(o *Options) {
		o.Parallel = n
	}
}

// SetTaskHandlerFunc set the task handler function
func SetTaskHandlerFunc(fn MessageHandleFunc) Option {
	return func(o *Options) {
		o.HandleFunc = fn
	}
}

// SetTaskTimeout set the timeout interval for the task handle function
// If task running time exceed the task time, ErrTimeout error is raise
func SetTaskTimeout(t time.Duration) Option {
	return func(o *Options) {
		o.TaskTimeout = t
	}
}

// SetParseRequestFunc set the function to parse request
func SetParseRequestFunc(fn ParseRequestFunc) Option {
	return func(o *Options) {
		o.ParseRequestFunc = fn
	}
}

// SetLogger set agent logger
func SetLogger(l Logger) Option {
	return func(o *Options) {
		o.Logger = l
	}
}
