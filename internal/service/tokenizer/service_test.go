package tokenizer

import (
	"context"
	"testing"

	"github.com/monobearotaku/online-chat-api/internal/domain/token"
	"github.com/monobearotaku/online-chat-api/internal/domain/token/data"
	"github.com/stretchr/testify/assert"
)

func Test_tokenizer_ValidateToken(t *testing.T) {
	t.Parallel()

	tr := NewTokenizer()
	ctx := context.Background()
	tkn, _ := tr.CreateToken(ctx, data.TokenData{})

	tests := []struct {
		name  string
		token token.Token
		err   error
	}{
		{
			name:  "Valid Token",
			token: tkn,
			err:   nil,
		},
		{
			name:  "Invalid Token",
			token: "",
			err:   token.ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tr.ValidateToken(ctx, tt.token)
			assert.Equal(t, tt.err, err)
		})
	}
}
