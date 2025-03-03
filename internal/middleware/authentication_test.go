package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"

	"gosl/pkg/contexts"
	"gosl/pkg/db"
	"gosl/pkg/tests"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticationMiddleware(t *testing.T) {
	cfg, err := tests.TestConfig()
	require.NoError(t, err)
	logger := tests.NilLogger()
	ver, err := strconv.ParseInt(cfg.DBName, 10, 0)
	require.NoError(t, err)
	wconn, rconn, err := tests.SetupTestDB(ver)
	require.NoError(t, err)
	sconn := db.MakeSafe(wconn, rconn, logger)
	defer sconn.Close()

	// Handler to check outcome of Authentication middleware
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := contexts.GetUser(r.Context())
		if user == nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(strconv.Itoa(0)))
			return
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(strconv.Itoa(user.ID)))
		}
	})
	var maint uint32
	atomic.StoreUint32(&maint, 0)
	// Add the middleware and create the server
	authHandler := Authentication(logger, cfg, sconn, testHandler, &maint)
	require.NoError(t, err)
	server := httptest.NewServer(authHandler)
	defer server.Close()

	tokens := getTokens()

	tests := []struct {
		name         string
		id           int
		accessToken  string
		refreshToken string
		expectedCode int
	}{
		{
			name:         "Valid Access Token (Fresh)",
			id:           1,
			accessToken:  tokens["accessFresh"],
			refreshToken: "",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Valid Access Token (Unfresh)",
			id:           1,
			accessToken:  tokens["accessUnfresh"],
			refreshToken: tokens["refreshExpired"],
			expectedCode: http.StatusOK,
		},
		{
			name:         "Valid Refresh Token (Triggers Refresh)",
			id:           1,
			accessToken:  tokens["accessExpired"],
			refreshToken: tokens["refreshValid"],
			expectedCode: http.StatusOK,
		},
		{
			name:         "Both tokens expired",
			accessToken:  tokens["accessExpired"],
			refreshToken: tokens["refreshExpired"],
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "Access token revoked",
			accessToken:  tokens["accessRevoked"],
			refreshToken: "",
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "Refresh token revoked",
			accessToken:  "",
			refreshToken: tokens["refreshRevoked"],
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "Invalid Tokens",
			accessToken:  tokens["invalid"],
			refreshToken: tokens["invalid"],
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "No Tokens",
			accessToken:  "",
			refreshToken: "",
			expectedCode: http.StatusUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &http.Client{}

			req, _ := http.NewRequest(http.MethodGet, server.URL, nil)

			// Add cookies if provided
			if tt.accessToken != "" {
				req.AddCookie(&http.Cookie{Name: "access", Value: tt.accessToken})
			}
			if tt.refreshToken != "" {
				req.AddCookie(&http.Cookie{Name: "refresh", Value: tt.refreshToken})
			}

			resp, err := client.Do(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCode, resp.StatusCode)
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, strconv.Itoa(tt.id), string(body))
		})
	}
}

// get the tokens to test with
func getTokens() map[string]string {
	tokens := map[string]string{
		"accessFresh":    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjQ4OTU2NzIyMTAsImZyZXNoIjo0ODk1NjcyMjEwLCJpYXQiOjE3Mzk2NzIyMTAsImlzcyI6IjEyNy4wLjAuMSIsImp0aSI6ImE4Njk2YWM4LTg3OWMtNDdkNC1iZWM2LTRlY2Y4MTRiZThiZiIsInNjb3BlIjoiYWNjZXNzIiwic3ViIjoxLCJ0dGwiOiJzZXNzaW9uIn0.6nAquDY0JBLPdaJ9q_sMpKj1ISG4Vt2U05J57aoPue8",
		"accessUnfresh":  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjMzMjk5Njc1NjcxLCJmcmVzaCI6MTczOTY3NTY3MSwiaWF0IjoxNzM5Njc1NjcxLCJpc3MiOiIxMjcuMC4wLjEiLCJqdGkiOiJjOGNhZmFjNy0yODkzLTQzNzMtOTI4ZS03MGUwODJkYmM2MGIiLCJzY29wZSI6ImFjY2VzcyIsInN1YiI6MSwidHRsIjoic2Vzc2lvbiJ9.plWQVFwHlhXUYI5utS7ny1JfXjJSFrigkq-PnTHD5VY",
		"accessExpired":  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3Mzk2NzIyNDgsImZyZXNoIjoxNzM5NjcyMjQ4LCJpYXQiOjE3Mzk2NzIyNDgsImlzcyI6IjEyNy4wLjAuMSIsImp0aSI6IjgxYzA1YzBjLTJhOGItNGQ2MC04Yzc4LWY2ZTQxODYxZDFmNCIsInNjb3BlIjoiYWNjZXNzIiwic3ViIjoxLCJ0dGwiOiJzZXNzaW9uIn0.iI1f17kKTuFDEMEYltJRIwRYgYQ-_nF9Wsn0KR6x77Q",
		"refreshValid":   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjQ4OTU2NzE5MjIsImlhdCI6MTczOTY3MTkyMiwiaXNzIjoiMTI3LjAuMC4xIiwianRpIjoiZTUxMTY3ZWEtNDA3OS00ZTczLTkzZDQtNTgwZDMzODRjZDU4Iiwic2NvcGUiOiJyZWZyZXNoIiwic3ViIjoxLCJ0dGwiOiJzZXNzaW9uIn0.tvtqQ8Z4WrYWHHb0MaEPdsU2FT2KLRE1zHOv3ipoFyc",
		"refreshExpired": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3Mzk2NzIyNDgsImlhdCI6MTczOTY3MjI0OCwiaXNzIjoiMTI3LjAuMC4xIiwianRpIjoiZTg5YTc5MTYtZGEzYi00YmJhLWI3ZDMtOWI1N2ViNjRhMmU0Iiwic2NvcGUiOiJyZWZyZXNoIiwic3ViIjoxLCJ0dGwiOiJzZXNzaW9uIn0.rH_fytC7Duxo598xacu820pQKF9ELbG8674h_bK_c4I",
		"accessRevoked":  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjQ4OTU2NzE5MjIsImZyZXNoIjoxNzM5NjcxOTIyLCJpYXQiOjE3Mzk2NzE5MjIsImlzcyI6IjEyNy4wLjAuMSIsImp0aSI6IjBhNmIzMzhlLTkzMGEtNDNmZS04ZjcwLTFhNmRhZWQyNTZmYSIsInNjb3BlIjoiYWNjZXNzIiwic3ViIjoxLCJ0dGwiOiJzZXNzaW9uIn0.mZLuCp9amcm2_CqYvbHPlk86nfiuy_Or8TlntUCw4Qs",
		"refreshRevoked": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjMzMjk5Njc1NjcxLCJpYXQiOjE3Mzk2NzU2NzEsImlzcyI6IjEyNy4wLjAuMSIsImp0aSI6ImI3ZmE1MWRjLTg1MzItNDJlMS04NzU2LTVkMjViZmIyMDAzYSIsInNjb3BlIjoicmVmcmVzaCIsInN1YiI6MSwidHRsIjoic2Vzc2lvbiJ9.5Q9yDZN5FubfCWHclUUZEkJPOUHcOEpVpgcUK-ameHo",
		"invalid":        "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0ODUxNDA5ODQsImlhdCI6MTQ4NTEzNzM4NCwiaXNzIjoiYWNtZS5jb20iLCJzdWIiOiIyOWFjMGMxOC0wYjRhLTQyY2YtODJmYy0wM2Q1NzAzMThhMWQiLCJhcHBsaWNhdGlvbklkIjoiNzkxMDM3MzQtOTdhYi00ZDFhLWFmMzctZTAwNmQwNWQyOTUyIiwicm9sZXMiOltdfQ.Mp0Pcwsz5VECK11Kf2ZZNF_SMKu5CgBeLN9ZOP04kZo",
	}
	return tokens
}
