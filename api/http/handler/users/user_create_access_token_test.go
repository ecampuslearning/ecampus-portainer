package users

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/apikey"
	"github.com/portainer/portainer/api/datastore"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/internal/testhelpers"
	"github.com/portainer/portainer/api/jwt"

	"github.com/segmentio/encoding/json"
	"github.com/stretchr/testify/assert"
)

func Test_userCreateAccessToken(t *testing.T) {
	is := assert.New(t)

	_, store := datastore.MustNewTestStore(t, true, true)

	// create admin and standard user(s)
	adminUser := &portainer.User{ID: 1, Password: "password", Username: "admin", Role: portainer.AdministratorRole}
	err := store.User().Create(adminUser)
	is.NoError(err, "error creating admin user")

	user := &portainer.User{ID: 2, Username: "standard", Role: portainer.StandardUserRole}
	err = store.User().Create(user)
	is.NoError(err, "error creating user")

	// setup services
	jwtService, err := jwt.NewService("1h", store)
	is.NoError(err, "Error initiating jwt service")
	apiKeyService := apikey.NewAPIKeyService(store.APIKeyRepository(), store.User())
	requestBouncer := security.NewRequestBouncer(store, jwtService, apiKeyService)
	rateLimiter := security.NewRateLimiter(10, 1*time.Second, 1*time.Hour)
	passwordChecker := security.NewPasswordStrengthChecker(store.SettingsService)

	h := NewHandler(requestBouncer, rateLimiter, apiKeyService, passwordChecker)
	h.DataStore = store
	h.CryptoService = testhelpers.NewCryptoService()

	// generate standard and admin user tokens
	adminJWT, _, _ := jwtService.GenerateToken(&portainer.TokenData{ID: adminUser.ID, Username: adminUser.Username, Role: adminUser.Role})
	jwt, _, _ := jwtService.GenerateToken(&portainer.TokenData{ID: user.ID, Username: user.Username, Role: user.Role})

	t.Run("standard user successfully generates API key", func(t *testing.T) {
		data := userAccessTokenCreatePayload{Password: "password", Description: "test-token"}
		payload, err := json.Marshal(data)
		is.NoError(err)

		req := httptest.NewRequest(http.MethodPost, "/users/2/tokens", bytes.NewBuffer(payload))
		testhelpers.AddTestSecurityCookie(req, jwt)

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		is.Equal(http.StatusOK, rr.Code)

		body, err := io.ReadAll(rr.Body)
		is.NoError(err, "ReadAll should not return error")

		var resp accessTokenResponse
		err = json.Unmarshal(body, &resp)
		is.NoError(err, "response should be json")
		is.EqualValues(data.Description, resp.APIKey.Description)
		is.NotEmpty(resp.RawAPIKey)
	})

	t.Run("admin cannot generate API key for standard user", func(t *testing.T) {
		data := userAccessTokenCreatePayload{Password: "password", Description: "test-token-admin"}
		payload, err := json.Marshal(data)
		is.NoError(err)

		req := httptest.NewRequest(http.MethodPost, "/users/2/tokens", bytes.NewBuffer(payload))
		testhelpers.AddTestSecurityCookie(req, adminJWT)

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		is.Equal(http.StatusForbidden, rr.Code)

		_, err = io.ReadAll(rr.Body)
		is.NoError(err, "ReadAll should not return error")
	})

	t.Run("endpoint cannot generate api-key using api-key auth", func(t *testing.T) {
		rawAPIKey, _, err := apiKeyService.GenerateApiKey(*user, "test-api-key")
		is.NoError(err)

		data := userAccessTokenCreatePayload{Password: "password", Description: "test-token-fails"}
		payload, err := json.Marshal(data)
		is.NoError(err)

		req := httptest.NewRequest(http.MethodPost, "/users/2/tokens", bytes.NewBuffer(payload))
		req.Header.Add("x-api-key", rawAPIKey)

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		is.Equal(http.StatusUnauthorized, rr.Code)

		body, err := io.ReadAll(rr.Body)
		is.NoError(err, "ReadAll should not return error")
		is.Equal(`{"message":"Auth not supported","details":"Authentication required"}`, string(body))
	})
}

func Test_userAccessTokenCreatePayload(t *testing.T) {
	is := assert.New(t)

	tests := []struct {
		payload    userAccessTokenCreatePayload
		shouldFail bool
	}{
		{
			payload:    userAccessTokenCreatePayload{Password: "password", Description: "test-token"},
			shouldFail: false,
		},
		{
			payload:    userAccessTokenCreatePayload{Password: "password", Description: ""},
			shouldFail: true,
		},
		{
			payload:    userAccessTokenCreatePayload{Password: "password", Description: "test token"},
			shouldFail: false,
		},
		{
			payload:    userAccessTokenCreatePayload{Password: "password", Description: "test-token "},
			shouldFail: false,
		},
		{
			payload: userAccessTokenCreatePayload{Password: "password", Description: `
this string is longer than 128 characters and hence this will fail.
this string is longer than 128 characters and hence this will fail.
this string is longer than 128 characters and hence this will fail.
this string is longer than 128 characters and hence this will fail.
this string is longer than 128 characters and hence this will fail.
this string is longer than 128 characters and hence this will fail.
`},
			shouldFail: true,
		},
	}

	for _, test := range tests {
		err := test.payload.Validate(nil)
		if test.shouldFail {
			is.Error(err)
		} else {
			is.NoError(err)
		}
	}
}
