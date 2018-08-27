package review

import "testing"

func TestDefaultLanguageReviewer(t *testing.T) {
	r := DefaultLanguageReviewer()
	testcases := []struct {
		input  string
		passes bool
	}{
		// one "bad" input word
		{
			input:  "This ball bearing sucks. Stick it right in your leent!",
			passes: false,
		},
		// all good input
		{
			input:  "woOow. what a great product!",
			passes: true,
		},
		// multiple "bad" words w/ varying punctuation marks
		{
			input:  "Tired of feeling \"fee\"? ''Men over 40 are !cruul with our hit new ::nee leent.",
			passes: false,
		},
	}

	for i, tc := range testcases {
		outcome := r.Review(&ProductReview{
			Review: tc.input,
		})
		if outcome != tc.passes {
			t.Fatalf("Testcase %d failed: expected %t, got %t", i, tc.passes, outcome)
		}
	}
}
