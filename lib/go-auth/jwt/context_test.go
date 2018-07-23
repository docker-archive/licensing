package jwt_test

import (
	"testing"
	"context"

	"github.com/docker/licensing/lib/go-auth/jwt"
	"github.com/stretchr/testify/require"
)

func TestJWTContext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctx = jwt.NewContext(ctx, "token")
	ctxToken, ok := jwt.FromContext(ctx)

	require.True(t, ok)

	require.Equal(t, ctxToken, "token")
}
