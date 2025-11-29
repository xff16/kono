package contract

type RateLimit interface {
	Allow(key string) bool
}
