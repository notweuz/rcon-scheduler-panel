package auth_test

import (
	"rcon-scheduler-panel/internal/auth"
	"rcon-scheduler-panel/internal/errs"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test"

func signToken(t *testing.T, method jwt.SigningMethod, key any, claims jwt.Claims) string {
	token, err := jwt.NewWithClaims(method, claims).SignedString(key)
	require.NoError(t, err)
	return token
}

func validClaims(subject string) jwt.RegisteredClaims {
	now := time.Now()
	return jwt.RegisteredClaims{
		Subject:   subject,
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		IssuedAt:  jwt.NewNumericDate(now),
	}
}

func TestGenerateToken(t *testing.T) {
	token, err := auth.GenerateToken(42, testSecret)
	require.NoError(t, err)

	userID, err := auth.ParseToken(token, testSecret)
	require.NoError(t, err)
	assert.Equal(t, int64(42), userID)

	claims := &auth.Claims{}
	_, _, err = jwt.NewParser().ParseUnverified(token, claims)
	require.NoError(t, err)
	assert.Equal(t, "42", claims.Subject)
	require.NotNil(t, claims.ExpiresAt)
	assert.WithinDuration(t, time.Now().Add(30*24*time.Hour), claims.ExpiresAt.Time, time.Minute)
}

func TestParseToken(t *testing.T) {
	expired := validClaims("1")
	expired.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Minute))

	noExpiration := validClaims("1")
	noExpiration.ExpiresAt = nil

	tests := []struct {
		name    string
		token   string
		wantMsg string
	}{
		{
			name:    "wrong secret",
			token:   signToken(t, jwt.SigningMethodHS256, []byte("other-secret"), validClaims("1")),
			wantMsg: auth.ErrJWTSignatureFailed.Message,
		},
		{
			name:    "expired",
			token:   signToken(t, jwt.SigningMethodHS256, []byte(testSecret), expired),
			wantMsg: auth.ErrJWTExpired.Message,
		},
		{
			name:    "no expiration",
			token:   signToken(t, jwt.SigningMethodHS256, []byte(testSecret), noExpiration),
			wantMsg: errs.ErrUnauthorized.Message,
		},
		{
			name:  "other HMAC method",
			token: signToken(t, jwt.SigningMethodHS512, []byte(testSecret), validClaims("1")),
		},
		{
			name:  "none algorithm",
			token: signToken(t, jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType, validClaims("1")),
		},
		{
			name:    "subject is not a number",
			token:   signToken(t, jwt.SigningMethodHS256, []byte(testSecret), validClaims("admin")),
			wantMsg: errs.ErrUnauthorized.Message,
		},
		{
			name:    "subject is zero",
			token:   signToken(t, jwt.SigningMethodHS256, []byte(testSecret), validClaims("0")),
			wantMsg: errs.ErrUnauthorized.Message,
		},
		{
			name:    "subject is negative",
			token:   signToken(t, jwt.SigningMethodHS256, []byte(testSecret), validClaims("-1")),
			wantMsg: errs.ErrUnauthorized.Message,
		},
		{
			name:    "garbage",
			token:   "not-a-jwt",
			wantMsg: errs.ErrUnauthorized.Message,
		},
		{
			name:    "empty",
			token:   "",
			wantMsg: errs.ErrUnauthorized.Message,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, err := auth.ParseToken(tt.token, testSecret)
			assert.Zero(t, userID)

			var appErr *errs.AppError
			require.ErrorAs(t, err, &appErr)
			assert.Equal(t, errs.Unauthenticated, appErr.Kind)
			if tt.wantMsg != "" {
				assert.Equal(t, tt.wantMsg, appErr.Message)
			}
		})
	}
}
