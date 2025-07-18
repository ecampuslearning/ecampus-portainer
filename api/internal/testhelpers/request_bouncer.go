package testhelpers

import (
	"net/http"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/dataservices"
)

type testRequestBouncer struct{}

// NewTestRequestBouncer creates new mock for requestBouncer
func NewTestRequestBouncer() *testRequestBouncer {
	return &testRequestBouncer{}
}

func (testRequestBouncer) AuthenticatedAccess(h http.Handler) http.Handler {
	return h
}

func (testRequestBouncer) AdminAccess(h http.Handler) http.Handler {
	return h
}

func (testRequestBouncer) RestrictedAccess(h http.Handler) http.Handler {
	return h
}

func (testRequestBouncer) PublicAccess(h http.Handler) http.Handler {
	return h
}

func (testRequestBouncer) TeamLeaderAccess(h http.Handler) http.Handler {
	return h
}

func (testRequestBouncer) EdgeComputeOperation(h http.Handler) http.Handler {
	return h
}

func (testRequestBouncer) AuthorizedEndpointOperation(r *http.Request, endpoint *portainer.Endpoint) error {
	return nil
}

func (testRequestBouncer) AuthorizedEdgeEndpointOperation(r *http.Request, endpoint *portainer.Endpoint) error {
	return nil
}

func (testRequestBouncer) TrustedEdgeEnvironmentAccess(tx dataservices.DataStoreTx, endpoint *portainer.Endpoint) error {
	return nil
}

func (testRequestBouncer) CookieAuthLookup(r *http.Request) (*portainer.TokenData, error) {
	return nil, nil
}

func (testRequestBouncer) JWTAuthLookup(r *http.Request) (*portainer.TokenData, error) {
	return nil, nil
}

func (testRequestBouncer) RevokeJWT(jti string) {}

func (testRequestBouncer) DisableCSP() {}

// AddTestSecurityCookie adds a security cookie to the request
func AddTestSecurityCookie(r *http.Request, jwt string) {
	r.AddCookie(&http.Cookie{
		Name:  portainer.AuthCookieKey,
		Value: jwt,
	})
}
