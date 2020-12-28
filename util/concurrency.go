package util

type Limiter struct {
	tokens chan bool
	limit  int
}

func (l *Limiter) Go(f func()) {
	l.tokens <- true
	go func() {
		f()
		<-l.tokens
	}()
}

func (l *Limiter) Close() {
	for i := 0; i < l.limit; i++ {
		l.tokens <- true
	}
}

func NewLimiter(limit int) *Limiter {
	return &Limiter{
		tokens: make(chan bool, limit),
		limit:  limit,
	}
}
