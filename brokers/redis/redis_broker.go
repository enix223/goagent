package redis

import (
	"encoding/json"
	"log"

	"github.com/enix223/goagent"
	"github.com/go-redis/redis/v7"
)

// Broker redis broker
type Broker struct {
	url              string
	client           *redis.Client
	done             <-chan struct{}
	requestChannel   chan goagent.Request
	parseRequestFunc goagent.ParseRequestFunc
}

// NewBroker create redis broker
func NewBroker(opts ...Option) *Broker {
	b := &Broker{}

	for _, o := range opts {
		o(b)
	}

	uri, err := redis.ParseURL(b.url)
	if err != nil {
		panic(err)
	}
	b.client = redis.NewClient(uri)

	return b
}

// Subscribe subscribe specific task topic, and return the request channel
// if success, or return error if subscribe failed. The `done` channel is
// used to notify the broker that the agent will be stopped
func (b *Broker) Subscribe(agentID, requestTopic string, done <-chan struct{}, fn goagent.ParseRequestFunc) (<-chan goagent.Request, error) {
	pubsub := b.client.PSubscribe(requestTopic)
	b.requestChannel = make(chan goagent.Request, 0)
	b.done = done

	log.Printf("Waiting for task in topic: %s...", requestTopic)
	go func() {
		for {
			select {
			case <-done:
				// finished and close
				close(b.requestChannel)
				pubsub.Close()
				return
			case msg, ok := <-pubsub.Channel():
				if ok {
					if request, err := fn([]byte(msg.Payload)); err == nil {
						b.requestChannel <- request
					} else {
						// failed to parse request
						log.Printf("Failed to parse request: %v", err)
					}
				}
			}
		}
	}()

	return b.requestChannel, nil
}

// NotifyResult when the task handler finished processing the request,
// response will be send back to broker
func (b *Broker) NotifyResult(resultTopic string, requestID string, result goagent.Response) error {
	body, err := json.Marshal(&result)
	if err != nil {
		return err
	}
	res := b.client.Publish(resultTopic, body)
	return res.Err()
}
