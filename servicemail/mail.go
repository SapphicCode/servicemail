package servicemail

// Envelope represents a Mail object, with attached Servicemail metadata.
// The router attaches the received Mail and populates these fields.
type Envelope struct {
	Mail     Mail
	Platform string
	UserID   string
}

// Mail represents an email.
type Mail struct {
	SenderName string // sender name, for instance Google
	From       string // sanitized address, for instance noreply@google.com
	To         string // receiving address
	Subject    string // mail subject
	Body       string // mail body
}
