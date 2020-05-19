package rabbitmq

// Option config Broker option
type Option func(a *Broker)

// SetExchange set task MQ exchange
func SetExchange(ex string) Option {
	return func(a *Broker) {
		a.exchange = ex
	}
}

// SetURL set url
func SetURL(url string) Option {
	return func(a *Broker) {
		a.url = url
	}
}
