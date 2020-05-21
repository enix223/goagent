package redis

import "github.com/enix223/goagent"

// Option config Broker option
type Option func(a *Broker)

// SetURL set url
func SetURL(url string) Option {
	return func(a *Broker) {
		a.url = url
	}
}

// SetLogger set logger
func SetLogger(l goagent.Logger) Option {
	return func(a *Broker) {
		a.logger = l
	}
}
