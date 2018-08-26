package queue

import (
	"encoding/json"
	"fmt"
	"review_system/review"

	"github.com/gomodule/redigo/redis"
)

// ProductReviewJob represents a product review that needs processing
type ProductReviewJob struct {
	Review   review.ProductReview `json:"review"`
	Status   string               `json:"status"`
	Attempts int                  `json:"attempts"`
}

// WorkerPool is our simple wrapper around the redis connection pool
type WorkerPool struct {
	pool *redis.Pool
}

// NewWorkerPool returns a worker pool for communicating with redis
func NewWorkerPool(endpoint string, port int) *WorkerPool {
	return &WorkerPool{
		pool: newRedisPool(endpoint, port),
	}
}

// newRedisPool returns a pool of connections for connecting to Redis
func newRedisPool(endpoint string, port int) *redis.Pool {
	address := fmt.Sprint(endpoint, ":", port)
	return &redis.Pool{
		MaxIdle:   50,
		MaxActive: 5000,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", address)
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}
}

// PushReview pushes the given product review to the given list
func (w *WorkerPool) PushReview(r review.ProductReview, listName string, status string,
	attempts int) (queueLength int64, err error) {
	job := ProductReviewJob{
		Review:   r,
		Status:   status,
		Attempts: attempts,
	}
	msg, err := json.Marshal(&job)
	if err != nil {
		return -1, err
	}

	c := w.pool.Get()
	defer c.Close()
	n, err := c.Do("LPUSH", listName, string(msg))
	return n.(int64), err
}

// ProcessNextReview attempts to process the next product review in the queue,
// incrementing the attempts counter on the job in the process. While the
// job is being processed, it sits in the processing queue.
//
// If the job fails and the attempts counter is below the threshold,
// the job is committed back to the request queue.
//
// If the job fails and the attempts counter exceeds the threshold,
// the job is discarded.
func (w *WorkerPool) ProcessNextReview(reqQName string, procQName string) (err error) {
	c := w.pool.Get()
	defer c.Close()
	msg, err := c.Do("RPOPLPUSH", reqQName, procQName)
	if err != nil {
		return err
	}

	// TODO: check for nil return in value?
	fmt.Println(msg)

	job := &ProductReviewJob{}
	err = json.Unmarshal([]byte(msg.(string)), job)
	if err != nil {
		return err
	}

	return nil
}
