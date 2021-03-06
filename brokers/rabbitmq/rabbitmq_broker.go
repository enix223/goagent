package rabbitmq

import (
	"encoding/json"
	"time"

	"github.com/enix223/goagent"
	"github.com/enix223/goagent/brokers/utils"
	"github.com/enix223/goagent/logger"
	"github.com/streadway/amqp"
)

const (
	reconnectInterval = 5
	defaultTopic      = "goagent.tasks"
	defaultExchange   = "/"
)

// Broker rabbitmq broker
type Broker struct {
	queueName        string
	topic            string
	url              string
	exchange         string
	logger           goagent.Logger
	conn             *amqp.Connection
	channel          *amqp.Channel
	channelClosed    chan *amqp.Error
	done             <-chan struct{}
	requestChannel   chan goagent.Request
	parseRequestFunc goagent.ParseRequestFunc
}

// NewBroker create rabbitmq broker
func NewBroker(opts ...Option) *Broker {
	a := &Broker{
		exchange: defaultExchange,
	}

	for _, o := range opts {
		o(a)
	}

	if a.logger == nil {
		a.logger = logger.NewLogger(
			logger.SetLevel(logger.INFO),
		)
	}

	return a
}

// Subscribe subscribe given task topic, and return request channel
func (a *Broker) Subscribe(agentID, taskTopic string, done <-chan struct{}, fn goagent.ParseRequestFunc) (<-chan goagent.Request, error) {
	a.done = done
	a.requestChannel = make(chan goagent.Request, 0)
	a.parseRequestFunc = fn
	a.topic = taskTopic
	a.queueName = agentID

	go func() {
		a.run()

		for {
			select {
			case <-a.done:
				// close the connection
				a.logger.Infof("Stopping client...")
				a.close()
				return
			case err := <-a.channelClosed:
				a.logger.Errorf("channel closed: %v", err)
				a.run()
			}
		}
	}()

	return a.requestChannel, nil
}

// NotifyResult when the task handler finished processing the request,
// response will be send back to broker
func (a *Broker) NotifyResult(resultTopic string, requestID string, result goagent.Response) error {
	var msg amqp.Publishing

	body, err := json.Marshal(&result)
	if err != nil {
		return err
	}

	msg.Body = body
	err = a.channel.Publish(
		a.exchange,
		resultTopic,
		false,
		false,
		msg,
	)
	return err
}

//
// Private section
//

func (a *Broker) subscribe() {
	topic := a.topic
	a.logger.Debugf("Try to subscribe topic: %s", topic)

	utils.WaitUntil(a.done, func() bool {
		a.logger.Debugf("Declaring queue %s...", a.queueName)
		_, err := a.channel.QueueDeclare(
			a.queueName, // queue name
			true,        // durable
			false,       // delete when usused
			false,       // exclusive
			false,       // no-utils.WaitUntil
			nil,         // agrs
		)
		if err != nil {
			a.logger.Errorf("failed declare queue: %v, will try to reconnect after %d seconds...", err, reconnectInterval)
			time.Sleep(reconnectInterval * time.Second)
			return false
		}
		a.logger.Debugf("Queue declared")
		return true
	})

	utils.WaitUntil(a.done, func() bool {
		a.logger.Debugf("Binding queue: %s, topic: %s, exchange: %s...", a.queueName, topic, a.exchange)
		err := a.channel.QueueBind(
			a.queueName, // queue name
			topic,       // topic
			a.exchange,  // exchange
			false,       // no-utils.WaitUntil
			nil,
		)
		if err != nil {
			a.logger.Errorf("failed to bind queue: %v, will try to reconnect after %d seconds...", err, reconnectInterval)
			time.Sleep(reconnectInterval * time.Second)
			return false
		}
		a.logger.Debugf("Queue binded")
		return true
	})

	var msgs <-chan amqp.Delivery
	utils.WaitUntil(a.done, func() bool {
		var err error
		msgs, err = a.channel.Consume(
			a.queueName, // queue name
			"",          // consumer
			true,        // auto-ack
			true,        // exclusive
			false,       // no-local
			false,       // no-utils.WaitUntil
			nil,         // args
		)
		if err != nil {
			a.logger.Errorf("failed to consume msg: %v, will try after %d seconds...", err, reconnectInterval)
			time.Sleep(reconnectInterval * time.Second)
			return false
		}
		return true
	})

	a.logger.Infof("message subscribed")

	for {
		select {
		case msg, ok := <-msgs:
			if ok {
				if req, err := a.parseRequestFunc(msg.Body); err == nil {
					a.logger.Debugf("Got task request: %s", string(msg.Body))
					a.requestChannel <- req
				} else {
					a.logger.Errorf("Failed to parse request: %v", err)
				}
			}
		case <-a.channelClosed:
			a.logger.Errorf("message handler exit coz channel closed")
			return
		case <-a.done:
			a.channel.QueueUnbind(a.queueName, topic, a.exchange, nil)
			a.channel.QueueDelete(a.queueName, false, false, false)
			a.logger.Infof("clear subscription")
			return
		}
	}
}

// Connect create connection
func (a *Broker) connect() {
	utils.WaitUntil(a.done, func() bool {
		a.logger.Infof("Try to connect MQ: %s", a.url)
		conn, err := amqp.Dial(a.url)
		if err != nil {
			a.logger.Errorf("failed to connect MQ: %v, will try after %d seconds...", err, reconnectInterval)
			time.Sleep(reconnectInterval * time.Second)
			return false
		}
		a.conn = conn
		return true
	})

	utils.WaitUntil(a.done, func() bool {
		a.logger.Debugf("Creating channel...")
		channel, err := a.conn.Channel()
		if err != nil {
			a.logger.Errorf("failed to create channel: %v, will try after %d seconds", err, reconnectInterval)
			time.Sleep(reconnectInterval * time.Second)
			return false
		}
		a.channel = channel
		return true
	})

	a.channel.NotifyClose(a.channelClosed)
	a.logger.Infof("Connection created")
}

// Close close connection
func (a *Broker) close() {
	a.logger.Infof("Closing rabbitmq client")
	close(a.requestChannel)

	if a.channel != nil {
		a.channel.Close()
	}
	if a.conn != nil {
		a.conn.Close()
	}
}

func (a *Broker) run() {
	a.channelClosed = make(chan *amqp.Error)
	select {
	case <-a.done:
		return
	default:
		a.connect()
		a.subscribe()
	}
}
