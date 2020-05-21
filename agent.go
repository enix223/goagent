package goagent

import (
	"context"
	"errors"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/enix223/goagent/logger"
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

	a.tokens = make(chan struct{}, a.opts.Parallel)
	for i := 0; i < a.opts.Parallel; i++ {
		// fill channel with tokens
		a.tokens <- struct{}{}
	}

	if a.opts.Logger == nil {
		a.opts.Logger = logger.NewLogger(
			logger.SetLevel(logger.INFO),
		)
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
	a.opts.Logger.Infof("Got reqeust: %s", request.GetID())
	a.opts.Logger.Debugf("Accquiring token...")

	// accquire token before action
	select {
	case <-a.tokens:
		// wait for token
		break
	case <-a.closeCh:
		// finished
		return
	}

	a.opts.Logger.Debugf("Token accquired")

	ctx, cancel := context.WithTimeout(context.Background(), a.opts.TaskTimeout)

	// prevent crash
	defer func() {
		a.opts.Logger.Debugf("Token released")

		// release token after task done
		a.tokens <- struct{}{}

		cancel()

		if err := recover(); err != nil {
			a.opts.Logger.Errorf("Error occurred")
			debug.PrintStack()
		}
	}()

	responseChannel := a.opts.HandleFunc(ctx, request)
	select {
	case response := <-responseChannel:
		// task request arrived
		a.opts.Logger.Debugf("Got task handler response, send result notification")
		a.opts.Broker.NotifyResult(a.opts.TaskResultTopic, request.GetID(), response)
	case <-a.closeCh:
		// agent stop
		a.opts.Logger.Infof("Agent stop")
		return
	case <-ctx.Done():
		// task timeout
		a.opts.Logger.Errorf("Task handle timeout. Send error notification")
		response := Response{
			RequestID: request.GetID(),
			Success:   false,
			Error:     ErrTimeout.Error(),
		}
		a.opts.Broker.NotifyResult(a.opts.TaskResultTopic, request.GetID(), response)
		return
	}
}
