package client

// Credentials contains one App credential. Secret authenticates the App and
// never substitutes for user wallet authorization.
type Credentials struct {
	// AppID is the partner application and environment identifier.
	AppID string
	// KeyID selects an active or rotation-grace credential.
	KeyID string
	// Secret is the credential secret shown only when created or rotated.
	Secret string
}

// Config contains immutable SDK configuration.
type Config struct {
	// BaseURL is the Pay API origin, normally https://pay-api.dextri.com.
	BaseURL string
	// Credentials authenticates one App in one environment.
	Credentials Credentials
	// UserAgent optionally identifies the integrating service and version.
	UserAgent string
}
