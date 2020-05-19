package rabbitmq_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/enix223/goagent"
	"github.com/enix223/goagent/brokers/rabbitmq"
)

type exmapleRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (e *exmapleRequest) GetID() string {
	return e.ID
}

func parseRequestFunc(body []byte) (goagent.Request, error) {
	var req exmapleRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, err
	}
	return &req, nil
}

func handleTask(ctx context.Context, request goagent.Request) <-chan goagent.Response {
	res := make(chan goagent.Response, 0)

	go func() {
		select {
		case <-ctx.Done():
			close(res)
			return
		default:
			// process logic goes here
			time.Sleep(1 * time.Second)
			req := request.(*exmapleRequest)
			response := goagent.Response{
				RequestID: req.ID,
				Success:   true,
				Result:    "Hi " + req.Name,
			}
			log.Println("Response: ", response)
			res <- response
		}
	}()

	return res
}

// ExampleBroker a task agent base on rabbitmq broker
// The agent can only process one request at a time
func ExampleBroker() {
	b := rabbitmq.NewBroker(
		rabbitmq.SetExchange("tasks"),
		rabbitmq.SetURL("amqp://guest:guest@localhost:5672"),
	)

	agent := goagent.NewAgent(
		goagent.SetAgentID("goagent.1"),                        // agent id
		goagent.SetBroker(b),                                   // rabbitmq broker
		goagent.SetParallel(1),                                 // only process one request at a time
		goagent.SetTaskRequestTopic("goagent.tasks.1.request"), // reqeust topic
		goagent.SetTaskResultTopic("goagent.tasks.1.result"),   // result topic
		goagent.SetTaskTimeout(time.Second*10),                 // processing timeout
		goagent.SetParseRequestFunc(parseRequestFunc),          // request parse function
		goagent.SetTaskHandlerFunc(handleTask),                 // function to process the request
	)

	done := make(chan os.Signal, 1)
	go func() {
		agent.Run()
	}()

	signal.Notify(done, os.Interrupt, os.Kill)
	<-done
	agent.Stop()
	fmt.Println("agent shutdown success")
	// Output: agent shutdown success
}
