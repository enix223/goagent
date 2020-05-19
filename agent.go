package goagent

import (
	"context"
	"errors"
	"log"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrTimeout taks timeout error
	ErrTimeout = errors.New("Task handle reach max running times")
)

const (
	defaultTaskTimeout = 15 * time.Second
)

// Agent task agent
type Agent struct {
	opts    Options
	closeCh chan struct{}
	tokens  chan struct{}
}

// NewAgent create task agent
func NewAgent(opts ...Option) *Agent {
	a := &Agent{
		opts:    Options{},
		closeCh: make(chan struct{}, 0),
	}

	for _, o := range opts {
		o(&a.opts)
	}

	maxParallel := runtime.NumCPU()
	if a.opts.Parallel <= 0 || a.opts.Parallel > maxParallel {
		a.opts.Parallel = maxParallel
	}

	if len(a.opts.ID) == 0 {
		a.opts.ID = uuid.New().String()
	}

	if a.opts.TaskTimeout == 0 {
		a.opts.TaskTimeout = defaultTaskTimeout
	}

	if a.opts.ParseRequestFunc == nil {
		panic("ParseRequestFunc not set")
	}

	a.tokens = make(chan struct{}, maxParallel)
	for i := 0; i < maxParallel; i++ {
		// fill channel with tokens
		a.tokens <- struct{}{}
	}

	return a
}

// Run run the task agent
func (a *Agent) Run() error {
	requestChannel, err := a.opts.Broker.Subscribe(a.opts.ID, a.opts.TaskRequestTopic, a.closeCh, a.opts.ParseRequestFunc)
	if err != nil {
		return err
	}

	for {
		select {
		case <-a.closeCh:
			return nil
		case request := <-requestChannel:
			go a.handleRequest(request)
		}
	}
}

// Stop stop the agent
func (a *Agent) Stop() {
	close(a.closeCh)
}

func (a *Agent) handleRequest(request Request) {
	log.Println("Start handling reqeust: ", request)

	// accquire token before action
	select {
	case <-a.tokens:
		// wait for token
		break
	case <-a.closeCh:
		// finished
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), a.opts.TaskTimeout)

	// prevent crash
	defer func() {
		// release token after task done
		a.tokens <- struct{}{}

		cancel()

		if err := recover(); err != nil {
			debug.PrintStack()
		}
	}()

	responseChannel := a.opts.HandleFunc(ctx, request)
	select {
	case response := <-responseChannel:
		// task request arrived
		a.opts.Broker.NotifyResult(a.opts.TaskResultTopic, request.GetID(), response)
	case <-a.closeCh:
		// agent stop
		return
	case <-ctx.Done():
		// task timeout
		response := Response{
			RequestID: request.GetID(),
			Success:   false,
			Error:     ErrTimeout.Error(),
		}
		a.opts.Broker.NotifyResult(a.opts.TaskResultTopic, request.GetID(), response)
		return
	}
}
