package identity_test

import (
	"context"
	"testing"

	"github.com/docker/licensing/lib/go-auth/identity"
	"github.com/stretchr/testify/require"
)

func defaultIdentity() identity.DockerIdentity {
	username := "testUsername"
	dockerID := "10557eca-6a92-4b97-8af2-f966572ac11e"
	email := "testEmail@gmail.com"

	return identity.DockerIdentity{
		Username: username,
		DockerID: dockerID,
		Email:    email,
	}
}

func TestIdentityContext(t *testing.T) {
	t.Parallel()

	dockerIdentity := defaultIdentity()

	ctx := context.Background()

	ctx = identity.NewContext(ctx, &dockerIdentity)
	ctxIdentity, ok := identity.FromContext(ctx)

	require.True(t, ok)

	require.Equal(t, dockerIdentity.Email, ctxIdentity.Email)
	require.Equal(t, dockerIdentity.Username, ctxIdentity.Username)
	require.Equal(t, dockerIdentity.DockerID, ctxIdentity.DockerID)
}

func TestDockerIdentity_String(t *testing.T) {
	t.Parallel()

	d := defaultIdentity()

	require.NotEmpty(t, d.String())
}
