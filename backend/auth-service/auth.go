package auth

import "encore.dev/beta/auth"

// ValidateToken is the Encore auth handler.
// It validates the Keycloak JWT and returns the user identity.
//
//encore:authhandler
func ValidateToken(ctx context.Context, token string) (auth.UID, *UserData, error) {
	// TODO: Phase 2 - validate Keycloak JWT
	panic("not implemented")
}

type UserData struct {
	Username string
	Email    string
	Roles    []string
}
