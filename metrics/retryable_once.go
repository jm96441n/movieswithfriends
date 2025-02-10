package metrics

import "sync"

type retryableOnce struct {
	once *sync.Once
	err  error
}

func NewRetryableOnce() *retryableOnce {
	return &retryableOnce{
		once: new(sync.Once),
	}
}

func (r *retryableOnce) Do(f func() error) error {
	// If we got an error last time, create a new sync.Once
	if r.err != nil {
		r.once = new(sync.Once)
		r.err = nil
	}

	r.once.Do(func() {
		r.err = f()
	})

	return r.err
}
