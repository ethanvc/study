package golangbreaker

type Breaker interface {
	Break() bool
	Name() string
}

type nullBreaker struct{}

func NewNullBreaker() Breaker {
	return nullBreaker{}
}

func (b nullBreaker) Name() string {
	return "null_breaker"
}

func (nullBreaker) Break() bool {
	return false
}
