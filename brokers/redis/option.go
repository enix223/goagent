package redis

// Option config Broker option
type Option func(a *Broker)

// SetURL set url
func SetURL(url string) Option {
	return func(a *Broker) {
		a.url = url
	}
}
