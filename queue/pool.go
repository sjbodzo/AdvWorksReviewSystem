package queue

import (
	"encoding/json"
	"fmt"

	"github.com/gomodule/redigo/redis"
	"github.com/sjbodzo/review_system/review"
)

// maxAttempts is the max number of times to try processing a product review job
// before discarding it and denying the client approval
var maxAttempts = 1

// ProductReviewJob represents a product review that needs processing
type ProductReviewJob struct {
	Review   review.ProductReview `json:"review"`
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
func (w *WorkerPool) PushReview(r review.ProductReview, listName string, attempts int) (queueLength int64, err error) {
	job := ProductReviewJob{
		Review:   r,
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
// job is being processed, it sits in the toQueue (processing) queue.
//
// If the job fails and the attempts counter is below the threshold,
// the job is committed back to the fromQueue (request) queue.
//
// If the job fails and the attempts counter exceeds the threshold,
// the job is discarded.
func (w *WorkerPool) ProcessNextReview(fromQueue string, toQueue string) (err error) {
	c := w.pool.Get()
	defer c.Close()
	msg, err := c.Do("RPOPLPUSH", fromQueue, toQueue)
	if err != nil {
		return err
	} else if msg == nil {
		return
	}

	var job ProductReviewJob
	err = json.Unmarshal(msg.([]byte), &job)
	if err != nil {
		return err
	}

	fmt.Println("job popped:", job)
	reviewer := review.DefaultLanguageReviewer()
	approved := job.Review.ApproveReview(reviewer)
	notifier := review.DefaultApprovalStatusNotifier()
	if approved {
		job.Review.NotifyClient("We hope to see you again soon!", true, notifier)
		err = w.RemoveReview(&job, toQueue)
		if err != nil {
			return err
		}
	} else if job.Attempts+1 >= maxAttempts {
		job.Review.NotifyClient("Please revise and resubmit your review!", true, notifier)
		err = w.RemoveReview(&job, toQueue)
		if err != nil {
			return err
		}
	} else {
		job.Attempts++
		_, err := c.Do("RPOPLPUSH", toQueue, fromQueue)
		if err != nil {
			return fmt.Errorf("Unable to re-queue job for review\nError: %v", err)
		}
	}

	return
}

// RemoveReview attempts to remove a review job, without queueing it anywhere else
func (w *WorkerPool) RemoveReview(job *ProductReviewJob, fromQueue string) (err error) {
	b, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("Unable to marshal review job\nError: %v", err)
	}

	c := w.pool.Get()
	defer c.Close()
	n, err := c.Do("LREM", fromQueue, 1, string(b))
	if err != nil {
		return fmt.Errorf("Unable to remove job from queue\nError: %v", err)
	} else if n.(int64) != 1 {
		return fmt.Errorf("Expected to remove %d jobs from queue, but removed %d instead", 1, n)
	}

	return nil
}
