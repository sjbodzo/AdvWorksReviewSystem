package review

import (
	"log"
	"regexp"
)

// Reviewer reviews the given ProductReview based on some criteria
type Reviewer interface {
	Review(pr *ProductReview) (approval bool)
}

// LanguageReviewer is a Reviewer that checks for blacklisted words
type LanguageReviewer struct {
	Blacklist  []string
	SplitRegex *regexp.Regexp
}

// NewLanguageReviewer returns a LanguageReviewer for use
func NewLanguageReviewer(blacklist []string, r *regexp.Regexp) *LanguageReviewer {
	return &LanguageReviewer{
		Blacklist:  blacklist,
		SplitRegex: r,
	}
}

// DefaultLanguageReviewer returns a default LanguageReviewer using sensible defaults
// for the blacklist of words and an effective regex to remove punctuation characters
func DefaultLanguageReviewer() *LanguageReviewer {
	blacklist := []string{"fee", "nee", "cruul", "leent"}
	r := regexp.MustCompile(`[.,\/#!$%\^&\*;:{}=\-_\x60~()\s]`)
	return NewLanguageReviewer(blacklist, r)
}

// Review ensures there is no blacklisted words present in the review's comment
func (l *LanguageReviewer) Review(pr *ProductReview) (approval bool) {
	for _, word := range l.SplitRegex.Split(pr.Review, -1) {
		for _, term := range l.Blacklist {
			if word == term {
				log.Printf("Review by %s denied approval due to usage"+
					" of blacklisted term\n", pr.EmailAddress)
				return false
			}
		}
	}
	return true
}
