package hashers

import "testing"

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
		_, _, _, _, _, err := ParseScryptStub(td.Password)
		if err != td.Error {
			t.Errorf("Expected %v for %s, got: %v", td.Error, td.Name, err)
		}
	}

}
