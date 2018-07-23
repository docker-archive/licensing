package jwt_test

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/docker/licensing/lib/go-auth/identity"
	"github.com/docker/licensing/lib/go-auth/jwt"

	"github.com/stretchr/testify/require"
)

func load(t *testing.T, fname string) []byte {
	b, err := ioutil.ReadFile(fname)
	require.NoError(t, err)
	return b
}

func defaultRootCertChain(t *testing.T) *x509.CertPool {
	rootCerts := load(t, "testdata/root-certs")

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(rootCerts)
	if !ok {
		t.Fatal("could not load root certs")
	}

	return roots
}

func defaultIdentity() identity.DockerIdentity {
	username := "testuser"
	dockerID := "00557eca-6a92-4b97-8af2-f966572ac11e"
	email := "testuser@gmail.com"

	return identity.DockerIdentity{
		Username: username,
		DockerID: dockerID,
		Email:    email,
		Scopes:   []string{"scopea", "scopeb"},
	}
}

func TestDecodeBadToken(t *testing.T) {
	t.Parallel()

	rootCerts := defaultRootCertChain(t)

	badToken := "foo"
	_, err := jwt.Decode(badToken, jwt.DecodeOptions{
		CertificateChain: rootCerts,
	})
	require.Error(t, err)
	require.Regexp(t, err, "malformed token error: token contains an invalid number of segments")
}

func TestDecodeBadCertChain(t *testing.T) {
	t.Parallel()

	privateKey := load(t, "testdata/private-key")
	trustedCert := load(t, "testdata/trusted-cert")

	// empty cert pool
	rootCerts := x509.NewCertPool()
	identity := defaultIdentity()

	tokenStr, err := jwt.Encode(identity, jwt.EncodeOptions{
		SigningKey:  privateKey,
		Certificate: trustedCert,
		Expiration:  time.Now().Add(time.Hour * 72).Unix(),
	})

	require.NoError(t, err)

	_, err = jwt.Decode(tokenStr, jwt.DecodeOptions{
		CertificateChain: rootCerts,
	})
	require.Error(t, err)
	require.Regexp(t, "certificate signed by unknown authority", err)
}

func TestExpiredToken(t *testing.T) {
	t.Parallel()

	privateKey := load(t, "testdata/private-key")
	trustedCert := load(t, "testdata/trusted-cert")

	rootCerts := defaultRootCertChain(t)
	identity := defaultIdentity()

	tokenStr, err := jwt.Encode(identity, jwt.EncodeOptions{
		SigningKey:  privateKey,
		Certificate: trustedCert,
		Expiration:  time.Now().Add(-(time.Hour * 72)).Unix(),
	})

	require.NoError(t, err)

	_, err = jwt.Decode(tokenStr, jwt.DecodeOptions{
		CertificateChain: rootCerts,
	})
	require.Error(t, err)
	require.Regexp(t, "token expiration error", err)
}

func TestIsExpired(t *testing.T) {
	t.Parallel()

	privateKey := load(t, "testdata/private-key")
	trustedCert := load(t, "testdata/trusted-cert")

	rootCerts := defaultRootCertChain(t)
	identity := defaultIdentity()

	// expired
	tokenStr, err := jwt.Encode(identity, jwt.EncodeOptions{
		SigningKey:  privateKey,
		Certificate: trustedCert,
		Expiration:  time.Now().Add(-(time.Hour * 72)).Unix(),
	})
	require.NoError(t, err)

	expired, err := jwt.IsExpired(tokenStr, jwt.DecodeOptions{
		CertificateChain: rootCerts,
	})
	require.NoError(t, err)
	require.True(t, expired)

	// not expired
	tokenStr, err = jwt.Encode(identity, jwt.EncodeOptions{
		SigningKey:  privateKey,
		Certificate: trustedCert,
		Expiration:  time.Now().Add(time.Hour * 72).Unix(),
	})
	require.NoError(t, err)

	expired, err = jwt.IsExpired(tokenStr, jwt.DecodeOptions{
		CertificateChain: rootCerts,
	})
	require.NoError(t, err)
	require.False(t, expired)
}

func TestEncodeExpectedFields(t *testing.T) {
	t.Parallel()

	privateKey := load(t, "testdata/private-key")
	trustedCert := load(t, "testdata/trusted-cert")

	identity := defaultIdentity()

	expiration := time.Now().Add(time.Hour * 72).Unix()

	tokenStr, err := jwt.Encode(identity, jwt.EncodeOptions{
		SigningKey:  privateKey,
		Certificate: trustedCert,
		Expiration:  expiration,
	})

	require.NoError(t, err)

	// manually parse the token
	tokenParts := strings.Split(tokenStr, ".")
	encodedClaims := tokenParts[1]

	claimsJSON, err := base64.StdEncoding.DecodeString(encodedClaims)
	require.NoError(t, err)

	var claims map[string]interface{}
	err = json.Unmarshal(claimsJSON, &claims)
	require.NoError(t, err)

	require.Equal(t, identity.Username, claims["username"])
	require.Equal(t, identity.DockerID, claims["sub"])
	require.Equal(t, identity.Email, claims["email"])
	require.Equal(t, strings.Join(identity.Scopes, " "), claims["scope"])

	var str string
	require.IsType(t, str, claims["jti"])
	require.NotEmpty(t, claims["jti"])

	var flt float64
	require.IsType(t, flt, claims["iat"])
	require.NotEmpty(t, claims["iat"])

	require.Equal(t, float64(expiration), claims["exp"])
}

func TestEncodeBadSigningKey(t *testing.T) {
	t.Parallel()

	privateKey := []byte("foo")

	trustedCert := load(t, "testdata/trusted-cert")

	identity := defaultIdentity()

	_, err := jwt.Encode(identity, jwt.EncodeOptions{
		SigningKey:  privateKey,
		Certificate: trustedCert,
		Expiration:  time.Now().Add(time.Hour * 72).Unix(),
	})

	require.Error(t, err)
	require.Regexp(t, "Invalid Key", err)
}

func TestTrustedToken(t *testing.T) {
	t.Parallel()

	privateKey := load(t, "testdata/private-key")
	trustedCert := load(t, "testdata/trusted-cert")

	rootCerts := defaultRootCertChain(t)
	identity := defaultIdentity()

	tokenStr, err := jwt.Encode(identity, jwt.EncodeOptions{
		SigningKey:  privateKey,
		Certificate: trustedCert,
		Expiration:  time.Now().Add(time.Hour * 72).Unix(),
	})

	require.NoError(t, err)

	decoded, err := jwt.Decode(tokenStr, jwt.DecodeOptions{
		CertificateChain: rootCerts,
	})
	require.NoError(t, err)

	require.Equal(t, identity.Email, decoded.Email)
	require.Equal(t, identity.Username, decoded.Username)
	require.Equal(t, identity.DockerID, decoded.DockerID)
	require.Equal(t, identity.Scopes, decoded.Scopes)
}

func TestUntrustedToken(t *testing.T) {
	t.Parallel()

	privateKey := load(t, "testdata/private-key")
	untrustedCert := load(t, "testdata/untrusted-cert")

	rootCerts := defaultRootCertChain(t)
	identity := defaultIdentity()

	tokenStr, err := jwt.Encode(identity, jwt.EncodeOptions{
		SigningKey:  privateKey,
		Certificate: untrustedCert,
		Expiration:  time.Now().Add(time.Hour * 72).Unix(),
	})

	require.NoError(t, err)

	_, err = jwt.Decode(tokenStr, jwt.DecodeOptions{
		CertificateChain: rootCerts,
	})
	require.Error(t, err)
	require.Regexp(t, "certificate signed by unknown authority", err)
}
