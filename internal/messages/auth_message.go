package messages

// DeviceCodeMsg carries the response from Keycloak's device authorization
// endpoint, or an error if the request failed.
type DeviceCodeMsg struct {
	DeviceCode              string
	UserCode                string
	VerificationURI         string
	VerificationURIComplete string
	ExpiresIn               int
	Interval                int
	Err                     error
}

// AuthPollTickMsg fires on a timer to trigger the next token poll.
type AuthPollTickMsg struct{}

// AuthResultMsg carries the outcome of a single token poll attempt.
// Pending polls (authorization_pending / slow_down) return a zero-value
// AuthResultMsg with Done false so the caller keeps waiting.
type AuthResultMsg struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
	Err          error
	Done         bool // true only on success or an unrecoverable error
}
