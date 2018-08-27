package review

import "log"

// ClientNotifier is a simple wrapper around anything that notifies a client
type ClientNotifier interface {
	Notify(p *ProductReview, approved bool, msg string) error
}

// ApprovalStatusNotifier notifies a client via Email on the status of their approval
// note: this is fake, but could include any sensible fields required to communicate
type ApprovalStatusNotifier struct {
	Sender string
}

// NewApprovalStatusNotifier returns a new NewApprovalStatusNotifier for notifying clients
func NewApprovalStatusNotifier(senderName string) *ApprovalStatusNotifier {
	return &ApprovalStatusNotifier{Sender: senderName}
}

// DefaultApprovalStatusNotifier provides a sensible default approval status notifier
func DefaultApprovalStatusNotifier() *ApprovalStatusNotifier {
	return &ApprovalStatusNotifier{Sender: "Bob"}
}

// Notify notifies the client of their approval status
func (notifier *ApprovalStatusNotifier) Notify(p *ProductReview, approved bool, msg string) error {
	s := "Hello, this is " + notifier.Sender + " from Foo Incorporated.\n"
	if approved {
		s += "Thank you for your review. It has been approved and will be on our site shortly!\n"
	} else {
		s += "Your review has been denied due to not meeting our corporate policies regarding language."
		s += "Please see our policies listed here: foo.inc/guidelines/community-practices.html\n"
	}
	s += msg
	log.Println("Notifying client:", s)
	return nil
}
