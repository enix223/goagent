# goagent

`goagent` is a library to help building task systems in golang. `goagent` currently supports `rabbitmq` and `redis` as MQ brokers.

## 1. Architecture

```
+---------+             +-------+------------+
| Agent.1 | <-------->  |       |            |
+---------+             |   g   |  RabbitMQ  |
                        |   o   |  Broker    |            Request
+---------+             |   a   |            | <---- TaskRequestTopic ----  +------------------+
| Agent.2 | <-------->  |   g   |------------+                              | Task distributor |
+---------+             |   e   |            | ----- TaskResultTopic  --->  +------------------+
                        |   n   |   Redis    |            Result
+---------+             |   t   |   Broker   |
| Agent.3 | <-------->  |       |            |
+---------+             +-------+------------+
```

## 2. Concepts

* ### Agent id

    Agent ID is an idenfitication for each task agent.

* ### Task request Topic

    Each task agent will subscribe the given task request topic. The task distributor will send task request to given task request topic for new task. `Task request topic` is not required to be unique accross different agents. Two or more agents can handle the same task request if you need to.

* ### Task result topic

    When the agent finishes processing the request, result is send back the task distributor through `Task result topic`. The result should be in form of [Result](./model.go).

* ### Parallel

    The parallel parameter is used to control how many tasks could be handled simultaneously. The upper boud is the number of cpus, the lower bound is 1. If parallel is set to 1, then only one task could be handled at a time, if the previous task not done, the succceeding task will be block until the previous one finished.

* ### Task timeout
  
    Task timeout is used to control how long the handler function is allowed to handle the request. If the task handle function exceed the timeout limit, `ErrTimeout` is raised.

* ### Request

    The request should conform to interface [Request](./model.go), which implement the `GetID` method. `GetID` should return the id for the request.

* ### Result

    The task handler should return [Response](./model.go) as task response. If `Success` field is `true`, means the task handler have process request successfully, otherwise, `Error` field will contain the error rease for the failure. `Result` field is the detail of the process result.

## 3. Examples

* For rabbitmq broker, please refer to [example_rabbitmq_test.go](./brokers/rabbitmq/example_rabbitmq_test.go)
* For redis broker, please refer to [example_redis_test.go](./brokers/redis/example_redis_test.go)
