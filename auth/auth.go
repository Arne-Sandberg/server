package auth

var provider CredentialsProvider

// Init intializes the auth package. You must call this before using any auth function.
func Init(cprovider CredentialsProvider) {
	provider = cprovider
	sessions = make(map[Session]struct{})
}

// Session represents a user session, denoted in a cryptographically secure string
type Session string

// TODO: we should probably store the sessions on the database, as re-logging every time the server
// restarts is kind of tedious.
var sessions map[Session]struct{}

// NewSession verifies the user's credentials and then returns a new Session
func NewSession(email string, password string) (Session, error) {
	validCredentials, err := provider.VerifyUserPassword(email, plaintext)
	if err != nil {
		log.Error(0, "Could not create new session, because call to credentials provider failed: %v", err)
		return nil, err
	}
	return utils.RandomStringSecure(15)
}

// ValidateSession checks if the session is valid.
func ValidateSession(sess Session) (valid bool, err error) {

}
