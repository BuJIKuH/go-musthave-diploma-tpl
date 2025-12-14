package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateTokenAndValidateToken(t *testing.T) {
	secret := "mysecret"
	userID := "user-123"

	tests := []struct {
		name         string
		modifySecret bool
		expired      bool
		wantErr      bool
	}{
		{
			name:         "#1 valid token",
			modifySecret: false,
			expired:      false,
			wantErr:      false,
		},
		{
			name:         "#2 wrong secret",
			modifySecret: true,
			expired:      false,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateToken(userID, secret)
			assert.NoError(t, err, "GenerateToken() should not return error")
			assert.NotEmpty(t, token, "GenerateToken() should return a token")

			validateSecret := secret
			if tt.modifySecret {
				validateSecret = "wrongsecret"
			}

			claims, err := ValidateToken(token, validateSecret)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, userID, claims.UserID)
				assert.WithinDuration(t, time.Now().Add(24*time.Hour), claims.ExpiresAt.Time, 5*time.Second)
			}
		})
	}
}
