package account

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"

	"github.com/qeek-dev/qeek-api-go-client/test"
)

const (
	// remember to change it to a valid token to run test
	// curl to get access token
	// curl /usr/local/bin/qcloud_auth_tool signin -i <client_id> -s <client-secret> -t password -u <user-name> -p <password>
	AccessToken = ""
)

type AccountTestCaseSuite struct {
	env     *test.Env
	service *Service
}

func setupAccountTestCase(t *testing.T) (AccountTestCaseSuite, func(t *testing.T)) {

	s := AccountTestCaseSuite{
		env: test.SetupEnv(t),
	}

	return s, func(t *testing.T) {

	}
}

func TestMeGetCall_Do(t *testing.T) {
	s, teardownTestCase := setupAccountTestCase(t)
	defer teardownTestCase(t)

	tt := []struct {
		name          string
		wantError     string
		setupTestCase test.SetupSubTest
	}{
		{
			name: "200",
			setupTestCase: func(t *testing.T) func(t *testing.T) {
				ctx := context.Background()
				ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: AccessToken})
				tc := oauth2.NewClient(ctx, ts)
				s.service = New(tc)

				return func(t *testing.T) {
				}
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			teardownSubTest := tc.setupTestCase(t)
			defer teardownSubTest(t)

			if AccessToken == "" {
				t.Skip("access token is empty, skip test")
			} else {
				res, err := s.service.Me.Get().Do()
				if err != nil {
					assert.Equal(t, err.Error(), tc.wantError, "An error was expected")
				} else {
					assert.NotNil(t, res)
				}
			}

		})
	}
}
