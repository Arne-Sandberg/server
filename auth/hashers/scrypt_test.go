package hashers

import (
	"testing"
)

func TestParseScryptStub(t *testing.T) {
	type data struct {
		Name     string
		Password string
		Error    error
	}

	var testdata = []data{
		{"empty password", "", ErrInvalidScryptStub},
		{"only preemble", "$s1$", ErrInvalidScryptStub},
		{"wrong preemble", "$s3$16384$8$1$SaltString$Base64Password=", ErrInvalidScryptStub},
		{"missing parts", "$s1$16384$8$1$", ErrInvalidScryptStub},
		{"valid password", "$s1$16384$8$1$VmFsaWRTYWx0$VmFsaWRQYXNzd29yZA0K", nil},
	}

	for _, td := range testdata {
		_, _, _, _, _, err := parseScryptStub(td.Password)
		if err != td.Error {
			t.Errorf("Expected %v for %s, got: %v", td.Error, td.Name, err)
		}
	}

}

func TestPasswordHashing(t *testing.T) {
	// verify that hashing the same password two times does not yield the same result
	h1, err := HashScrypt("testpassword")
	h2, err := HashScrypt("testpassword")
	if err != nil {
		t.Errorf("Error while hashing password: %v", err)
		return
	}
	if h1 == h2 {
		t.Errorf("Hashing the same password twice yielded the same result: %s", h1)
		return
	} else {
		t.Logf("h1: %s, h2: %s", h1, h2)
	}
	// first, hash a password, then test it against itself
	hash, err := HashScrypt("h4x0r!")
	if err != nil {
		t.Errorf("Got error while hashing password: %v", err)
		return
	}
	valid, err := ValidateScryptPassword("h4x0r!", hash)
	if err != nil {
		t.Errorf("Got error while validating password: %v", err)
		return
	}
	if !valid {
		t.Errorf("Expected a valid password verification, but it failed")
	}
}
