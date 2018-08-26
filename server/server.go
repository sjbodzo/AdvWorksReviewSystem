package server

import (
	"fmt"
	"net/http"
	"review_system/db"
	"review_system/queue"
	"time"
)

// New returns a new Server instance that can respond to requests to store reviews
func New(port int, version string, wrapper *db.Wrapper, pool *queue.WorkerPool) (*http.Server, error) {
	if wrapper == nil {
		return nil, fmt.Errorf("Server requires database to write to")
	}

	http.HandleFunc(fmt.Sprint("/", version, "/api/reviews"), ProductReview(wrapper, pool))

	srv := &http.Server{
		Handler:      http.DefaultServeMux,
		Addr:         fmt.Sprint(":", port),
		WriteTimeout: 1 * time.Second,
		ReadTimeout:  5 * time.Second,
	}
	return srv, nil
}
