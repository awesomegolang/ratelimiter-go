package ratelimiter_test

import (
	"testing"
	"time"

	"sync"

	"github.com/stretchr/testify/assert"
	ratelimiter "github.com/teambition/ratelimiter-go"
)

func TestRateLimiter(t *testing.T) {
	t.Run("ratelimiter with default Options should be", func(t *testing.T) {
		assert := assert.New(t)

		limiter := ratelimiter.New(ratelimiter.Options{})
		id := genID()
		policy := []int{10, 1000}

		res, err := limiter.Get(id, policy...)
		assert.Nil(err)
		assert.Equal(10, res.Total)
		assert.Equal(9, res.Remaining)
		assert.Equal(1000, int(res.Duration/time.Millisecond))
		assert.True(res.Reset.After(time.Now()))
		res, err = limiter.Get(id, policy...)
		assert.Equal(10, res.Total)
		assert.Equal(8, res.Remaining)
	})

	t.Run("ratelimiter with expire should be", func(t *testing.T) {
		assert := assert.New(t)

		limiter := ratelimiter.New(ratelimiter.Options{})
		id := genID()
		policy := []int{10, 100}

		res, err := limiter.Get(id, policy...)
		assert.Nil(err)
		assert.Equal(10, res.Total)
		assert.Equal(9, res.Remaining)
		res, err = limiter.Get(id, policy...)
		assert.Equal(8, res.Remaining)

		time.Sleep(100 * time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Nil(err)
		assert.Equal(10, res.Total)
		assert.Equal(9, res.Remaining)
	})

	t.Run("ratelimiter with goroutine should be", func(t *testing.T) {
		assert := assert.New(t)

		limiter := ratelimiter.New(ratelimiter.Options{})
		policy := []int{10, 500}
		id := genID()
		res, err := limiter.Get(id, policy...)
		assert.Nil(err)
		assert.Equal(10, res.Total)
		assert.Equal(9, res.Remaining)
		for i := 0; i < 100; i++ {
			go limiter.Get(id, policy...)
		}
		time.Sleep(200 * time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Nil(err)
		assert.Equal(10, res.Total)
		assert.Equal(-1, res.Remaining)
	})

	t.Run("ratelimiter with multi-policy should be", func(t *testing.T) {
		assert := assert.New(t)

		limiter := ratelimiter.New(ratelimiter.Options{})
		id := genID()
		policy := []int{3, 100, 2, 200}

		res, err := limiter.Get(id, policy...)
		assert.Nil(err)
		assert.Equal(3, res.Total)
		assert.Equal(2, res.Remaining)
		res, err = limiter.Get(id, policy...)
		assert.Equal(1, res.Remaining)
		res, err = limiter.Get(id, policy...)
		assert.Equal(0, res.Remaining)
		res, err = limiter.Get(id, policy...)
		assert.Equal(-1, res.Remaining)
		res, err = limiter.Get(id, policy...)
		assert.Equal(-1, res.Remaining)
		assert.True(res.Reset.After(time.Now()))

		time.Sleep(res.Duration + time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Equal(2, res.Total)
		assert.Equal(1, res.Remaining)
		assert.Equal(time.Millisecond*200, res.Duration)

		res, err = limiter.Get(id, policy...)
		assert.Equal(0, res.Remaining)
		res, err = limiter.Get(id, policy...)
		assert.Equal(-1, res.Remaining)

		time.Sleep(res.Duration + time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Equal(2, res.Total)
		assert.Equal(1, res.Remaining)
		assert.Equal(time.Millisecond*200, res.Duration)
	})

	t.Run("ratelimiter with Remove id should be", func(t *testing.T) {
		assert := assert.New(t)

		limiter := ratelimiter.New(ratelimiter.Options{})
		id := genID()
		policy := []int{10, 1000}

		res, err := limiter.Get(id, policy...)
		assert.Nil(err)
		assert.Equal(10, res.Total)
		assert.Equal(9, res.Remaining)
		limiter.Remove(id)
		res, err = limiter.Get(id, policy...)
		assert.Equal(10, res.Total)
		assert.Equal(9, res.Remaining)
	})

	t.Run("ratelimiter with wrong policy id should be", func(t *testing.T) {
		assert := assert.New(t)

		limiter := ratelimiter.New(ratelimiter.Options{})
		id := genID()
		policy := []int{10, 1000, 1}

		res, err := limiter.Get(id, policy...)
		assert.Error(err)
		assert.Equal(0, res.Total)
		assert.Equal(0, res.Remaining)
	})

	t.Run("ratelimiter with empty policy id should be", func(t *testing.T) {
		assert := assert.New(t)

		limiter := ratelimiter.New(ratelimiter.Options{})
		id := genID()
		policy := []int{}

		res, _ := limiter.Get(id, policy...)
		assert.Equal(100, res.Total)
		assert.Equal(99, res.Remaining)
		assert.Equal(time.Minute, res.Duration)
	})

	t.Run("limiter.Get with invalid args", func(t *testing.T) {
		assert := assert.New(t)

		limiter := ratelimiter.New(ratelimiter.Options{})
		id := genID()
		_, err := limiter.Get(id, 10)
		assert.Equal("ratelimiter: must be paired values", err.Error())

		_, err2 := limiter.Get(id, -1, 10)
		assert.Equal("ratelimiter: must be positive integer", err2.Error())

		_, err3 := limiter.Get(id, 10, 0)
		assert.Equal("ratelimiter: must be positive integer", err3.Error())
	})

	// t.Run("ratelimiter with Clean cache should be", func(t *testing.T) {
	// 	assert := assert.New(t)
	// 	limiter := ratelimiter.New(ratelimiter.Options{})

	// 	id := genID()
	// 	policy := []int{10, 100}

	// 	res, _ := limiter.Get(id, policy...)
	// 	assert.Equal(10, res.Total)
	// 	assert.Equal(9, res.Remaining)

	// 	time.Sleep(res.Duration + time.Millisecond)
	// 	limiter.Clean()
	// 	res, _ = limiter.Get(id, policy...)
	// 	assert.Equal(10, res.Total)
	// 	assert.Equal(9, res.Remaining)

	// 	time.Sleep(res.Duration + time.Millisecond)
	// 	limiter.Clean()
	// 	res, _ = limiter.Get(id, policy...)
	// 	assert.Equal(10, res.Total)
	// 	assert.Equal(9, res.Remaining)
	// })

	t.Run("ratelimiter with big goroutine should be", func(t *testing.T) {
		assert := assert.New(t)

		limiter := ratelimiter.New(ratelimiter.Options{})
		policy := []int{1000, 1000}
		id := genID()

		var wg sync.WaitGroup
		wg.Add(1000)
		for i := 0; i < 1000; i++ {
			go func() {
				newid := genID()
				limiter.Get(newid, policy...)
				limiter.Get(id, policy...)
				wg.Done()
			}()
		}
		wg.Wait()
		res, err := limiter.Get(id, policy...)
		assert.Nil(err)
		assert.Equal(1000, res.Total)
		assert.Equal(-1, res.Remaining)
	})

	// t.Run("ratelimiter with CleanDuration should be", func(t *testing.T) {
	// 	assert := assert.New(t)
	// 	limiter := ratelimiter.New(ratelimiter.Options{
	// 		CleanDuration: 100 * time.Millisecond,
	// 	})
	// 	policy := []int{100, 100}
	// 	id := genID()

	// 	res, err := limiter.Get(id, policy...)
	// 	assert.Nil(err)
	// 	assert.Equal(100, res.Total)
	// 	assert.Equal(99, res.Remaining)
	// 	time.Sleep(res.Duration + time.Millisecond)
	// 	res, err = limiter.Get(id, policy...)
	// 	assert.Equal(100, res.Total)
	// 	assert.Equal(99, res.Remaining)
	// })
	// t.Run("ratelimiter with CleanDuration should be", func(t *testing.T) {
	// 	assert := assert.New(t)
	// 	limiter := ratelimiter.New(ratelimiter.Options{})

	// 	limiter.Get("1", []int{100, 100}...)
	// 	limiter.Get("2", []int{100, 100}...)
	// 	assert.Equal(2, limiter.Count())
	// })

	t.Run("limiter.Get with multi-policy for expired", func(t *testing.T) {
		assert := assert.New(t)

		id := genID()
		policy := []int{2, 100, 2, 200, 1, 200, 1, 300}
		limiter := ratelimiter.New(ratelimiter.Options{})

		//First policy
		res, err := limiter.Get(id, policy...)
		assert.Nil(err)
		assert.Equal(2, res.Total)
		assert.Equal(1, res.Remaining)
		assert.Equal(time.Millisecond*100, res.Duration)

		res, err = limiter.Get(id, policy...)
		assert.Equal(0, res.Remaining)

		res, err = limiter.Get(id, policy...)
		assert.Equal(-1, res.Remaining)

		//Second policy
		time.Sleep(res.Duration + time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Equal(2, res.Total)
		assert.Equal(1, res.Remaining)
		assert.Equal(time.Millisecond*200, res.Duration)

		res, err = limiter.Get(id, policy...)
		assert.Equal(0, res.Remaining)

		res, err = limiter.Get(id, policy...)
		assert.Equal(-1, res.Remaining)

		//Third policy
		time.Sleep(res.Duration + time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Equal(1, res.Total)
		assert.Equal(0, res.Remaining)
		assert.Equal(time.Millisecond*200, res.Duration)

		res, err = limiter.Get(id, policy...)
		assert.Equal(-1, res.Remaining)

		// restore to First policy after Third policy*2 Duration
		time.Sleep(res.Duration*2 + time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Equal(2, res.Total)
		assert.Equal(1, res.Remaining)
		assert.Equal(time.Millisecond*100, res.Duration)

		//Second policy
		time.Sleep(res.Duration + time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Equal(2, res.Total)
		assert.Equal(1, res.Remaining)
		assert.Equal(time.Millisecond*200, res.Duration)

		res, err = limiter.Get(id, policy...)
		assert.Equal(0, res.Remaining)

		res, err = limiter.Get(id, policy...)
		assert.Equal(-1, res.Remaining)

		//Third policy
		time.Sleep(res.Duration + time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Equal(1, res.Total)
		assert.Equal(0, res.Remaining)
		assert.Equal(time.Millisecond*200, res.Duration)

		res, err = limiter.Get(id, policy...)
		assert.Equal(-1, res.Remaining)

		//Fourth policy
		time.Sleep(res.Duration + time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Equal(1, res.Total)
		assert.Equal(0, res.Remaining)
		assert.Equal(time.Millisecond*300, res.Duration)

		res, err = limiter.Get(id, policy...)
		assert.Equal(1, res.Total)
		assert.Equal(-1, res.Remaining)

		res, err = limiter.Get(id, policy...)
		assert.Equal(1, res.Total)
		assert.Equal(-1, res.Remaining)

		// restore to First policy after Fourth policy*2 Duration
		time.Sleep(res.Duration*2 + time.Millisecond)
		res, err = limiter.Get(id, policy...)
		assert.Equal(2, res.Total)
		assert.Equal(1, res.Remaining)
		assert.Equal(time.Millisecond*100, res.Duration)
	})
}