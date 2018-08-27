package review

import (
	"fmt"
	"html"
	"html/template"
	"regexp"
	"strings"
)

// ProductReview represents a client's product review
type ProductReview struct {
	ProductID    int    `json:"productid"`
	Review       string `json:"review"`
	ReviewerName string `json:"name"`
	EmailAddress string `json:"email"`
	Rating       int    `json:"rating"`
}

// Sanitize escapes html and javascript in the review, to help prevent XSS attacks
func (r *ProductReview) Sanitize() {
	r.Review = html.EscapeString(r.Review)
	r.Review = template.JSEscapeString(r.Review)
}

// NotifyClient notifies a client about their review with the given msg and notifiers
func (r *ProductReview) NotifyClient(msg string, approved bool, notifiers ...ClientNotifier) (errors []error) {
	for _, notifier := range notifiers {
		err := notifier.Notify(r, approved, msg)
		if err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

// ApproveReview vets the product review for approval using the passed in Reviewers
func (r *ProductReview) ApproveReview(reviewers ...Reviewer) bool {
	for _, reviewer := range reviewers {
		if reviewer.Review(r) == false {
			return false
		}
	}
	return true
}

// Validate ensures all the input values in the review are valid
func (r *ProductReview) Validate() (errors []error) {
	// check params exist
	var params []string
	if r.ProductID == 0 {
		params = append(params, "Product ID")
	}
	if r.Review == "" {
		params = append(params, "Review Text")
	}
	if r.ReviewerName == "" {
		params = append(params, "Reviewer Name")
	}
	if r.EmailAddress == "" {
		params = append(params, "Email Address")
	}
	if r.Rating == 0 {
		params = append(params, "Rating")
	}
	if params != nil {
		err := fmt.Errorf("Missing param(s): %s", strings.Join(params, ", "))
		errors = append(errors, err)
	}

	// validate reviewer name
	if r.ReviewerName != "" {
		if ok, _ := regexp.MatchString(`^[a-zA-Z0-9].$`, r.ReviewerName); ok {
			err := fmt.Errorf("Invalid reviewer name format: please use only characters and digits")
			errors = append(errors, err)
		}
	}

	// validate email address format
	if r.EmailAddress != "" {
		regex := regexp.MustCompile(`^[A-Za-z0-9._%-]+@[A-Za-z0-9.-]+[.][A-Za-z]+$`)
		if !regex.MatchString(r.EmailAddress) {
			err := fmt.Errorf("Invalid email address format")
			errors = append(errors, err)
		} else if len(r.EmailAddress) > 50 {
			err := fmt.Errorf("Email address exceeds max limit of 50 characters")
			errors = append(errors, err)
		}
	}

	// validate rating
	if r.Rating != 0 && (r.Rating > 5 || r.Rating < 1) {
		err := fmt.Errorf("Rating must be a value in the range of 1 to 5")
		errors = append(errors, err)
	}

	// validate comment, if any
	if r.Review != "" && len(r.Review) > 3850 {
		err := fmt.Errorf("Review length is limited to 3850 characters")
		errors = append(errors, err)
	}

	return errors
}
