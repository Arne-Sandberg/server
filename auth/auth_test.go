package auth

import (
	"testing"

	"github.com/riesinger/freecloud/models"
)

var credProvider CredentialsProvider

// testing implementation of a CredentialsProvider
type testCredProvider struct{}

func (p *testCredProvider) VerifyUserPassword(email string, plaintext string) (bool, error) {
	return true, nil
}

func (p *testCredProvider) NewUser(user models.User) error {
	return nil
}

func init() {
	credProvider = &testCredProvider{}
	Init(credProvider)
}

func TestNewSession(t *testing.T) {
	type pair struct {
		Name          string
		Email         string
		Password      string
		ExpectedValid bool
	}
	var testdata = []pair{
		{"valid email and password", "john.doe@test.com", "thisIsAPassword", true},
		{"missing email", "", "invalid cause no mail", false},
		{"missing password", "john.doe@test.com", "", false},
		{"missing email and password", "", "", false},
	}

	for _, td := range testdata {
		gotSession, gotError := NewSession(td.Email, td.Password)
		if td.ExpectedValid {
			if len(gotSession) != SessionTokenLength {
				t.Error("Expected a valid session for", td.Name, "but length of token is", len(gotSession))
			} else if gotError != nil {
				t.Error("Expected no error for", td.Name, "got", gotError)
			}
		} else {
			if len(gotSession) == SessionTokenLength {
				t.Error("Expected an empty session for", td.Name, "but length of token is", len(gotSession))
			} else if gotError == nil {
				t.Error("Expected an error for", td.Name, "but got none")
			}
		}
	}

}
