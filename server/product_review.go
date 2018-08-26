package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"review_system/db"
	"review_system/queue"
	"review_system/review"
)

// AddReviewResponse stores the response to the request
type AddReviewResponse struct {
	Success  bool     `json:"success"`
	ReviewID int      `json:"reviewID,omitempty"`
	Errors   []string `json:"errors,omitempty"`
}

// ProductReview is the handler for adding/updating product reviews
func ProductReview(db *db.Wrapper, pool *queue.WorkerPool) http.HandlerFunc {
	// fmtResponse formats the response as json for the client
	fmtResponse := func(reviewID *int, errors []error) string {
		var response AddReviewResponse
		if errors != nil {
			var errMsgs []string
			for _, err := range errors {
				errMsgs = append(errMsgs, err.Error())
			}
			response = AddReviewResponse{
				Success: false,
				Errors:  errMsgs,
			}
		} else {
			response = AddReviewResponse{
				Success:  true,
				ReviewID: *reviewID,
			}
		}

		m, err := json.Marshal(&response)
		if err != nil {
			panic(err)
		}
		return string(m)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			// if we can't decode it, just return a generic error
			decoder := json.NewDecoder(r.Body)
			var req review.ProductReview
			err := decoder.Decode(&req)
			if err != nil {
				log.Println(err)
				if err == io.EOF {
					err = fmt.Errorf("Request must include body")
				}
				http.Error(w, fmtResponse(nil, []error{err}), http.StatusBadRequest)
				return
			}

			// if there are specific input issues, return specific error(s)
			if errs := req.Validate(); errs != nil {
				http.Error(w, fmtResponse(nil, errs), http.StatusBadRequest)
				return
			}

			// request is valid, write it to the db & queue up for processing
			id, err := db.UpsertReview(req.ProductID, req.ReviewerName, req.EmailAddress, req.Rating, req.Review)
			if err != nil {
				log.Println(err) // log error, but hide it from the client
				http.Error(w, fmtResponse(nil, []error{fmt.Errorf("Server error")}), http.StatusBadRequest)
				return
			}
			go pool.PushReview(req, "req_queue", "PENDING", 0)

			m, _ := json.Marshal(&AddReviewResponse{
				ReviewID: id,
				Success:  true,
			})
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, string(m))
			return
		}
	}
}
