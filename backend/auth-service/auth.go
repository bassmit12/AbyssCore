package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"encore.dev/beta/auth"
	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

// UserData is injected into every authenticated request context.
type UserData struct {
	Subject string
	Email   string
	Roles   []string
}

var (
	jwksMu sync.Mutex
	jwks   keyfunc.Keyfunc
)

func getIssuer() string {
	return strings.TrimSuffix(os.Getenv("KEYCLOAK_ISSUER"), "/")
}

func getJWKS() (keyfunc.Keyfunc, error) {
	jwksMu.Lock()
	defer jwksMu.Unlock()
	if jwks != nil {
		return jwks, nil
	}
	url := os.Getenv("KEYCLOAK_JWKS_URL")
	if url == "" {
		issuer := getIssuer()
		if issuer == "" {
			return nil, errors.New("KEYCLOAK_ISSUER env var is required")
		}
		url = fmt.Sprintf("%s/protocol/openid-connect/certs", issuer)
	}
	k, err := keyfunc.NewDefaultCtx(context.Background(), []string{url})
	if err != nil {
		return nil, err
	}
	jwks = k
	return jwks, nil
}

func extractRoles(claims jwt.MapClaims) []string {
	roles := []string{}
	realm, ok := claims["realm_access"].(map[string]any)
	if !ok {
		return roles
	}
	rawRoles, ok := realm["roles"].([]any)
	if !ok {
		return roles
	}
	for _, r := range rawRoles {
		if s, ok := r.(string); ok {
			roles = append(roles, s)
		}
	}
	return roles
}

// ValidateToken is the Encore auth handler.
// It validates the Keycloak JWT from the Authorization header and returns the user identity.
//
//encore:authhandler
func ValidateToken(ctx context.Context, token string) (auth.UID, *UserData, error) {
	token = strings.TrimPrefix(strings.TrimSpace(token), "Bearer ")
	if token == "" {
		return "", nil, errors.New("missing token")
	}

	k, err := getJWKS()
	if err != nil {
		return "", nil, err
	}

	issuer := getIssuer()
	rawClaims := jwt.MapClaims{}
	_, err = jwt.ParseWithClaims(token, rawClaims, k.Keyfunc,
		jwt.WithIssuer(issuer),
		jwt.WithLeeway(30*time.Second),
	)
	if err != nil {
		return "", nil, fmt.Errorf("invalid token: %w", err)
	}

	sub, _ := rawClaims["sub"].(string)
	email, _ := rawClaims["email"].(string)

	return auth.UID(sub), &UserData{
		Subject: sub,
		Email:   email,
		Roles:   extractRoles(rawClaims),
	}, nil
}

// HasRole checks if the user has a specific Keycloak realm role.
func HasRole(ud *UserData, role string) bool {
	for _, r := range ud.Roles {
		if r == role {
			return true
		}
	}
	return false
}
