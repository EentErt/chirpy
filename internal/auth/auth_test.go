package auth

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

type makeJWTInput struct {
	id        uuid.UUID
	secret    string
	expiresIn time.Duration
}

type testValidateOutput struct {
	err1 error
	err2 error
}

func TestMakeJWT(t *testing.T) {
	testUUID, _ := uuid.NewUUID()
	testTime, _ := time.ParseDuration("10s")
	cases := []struct {
		input    makeJWTInput
		expected testValidateOutput
	}{
		{
			input: makeJWTInput{
				id:        testUUID,
				secret:    "test",
				expiresIn: testTime,
			},
			expected: testValidateOutput{
				err1: nil,
				err2: nil,
			},
		},
	}

	for _, c := range cases {
		testOut := testValidateOutput{}
		testToken, _ := MakeJWT(c.input.id, c.input.secret)
		_, err := ValidateJWT(testToken, c.input.secret)
		testOut.err1 = err
		time.Sleep(15 * time.Second)
		_, err = ValidateJWT(testToken, c.input.secret)
		testOut.err2 = err
		if testOut.err1 != c.expected.err1 {
			fmt.Println(err)
			t.Errorf("Validation Failed")
			continue
		} else if testOut.err2 == c.expected.err2 {
			t.Errorf("expiration failed")
			continue
		}
	}
}

func testGetBearerToken(t *testing.T) {
	cases := []struct {
		input    []string
		expected string
	}{
		{
			input:    []string{"Authorization", "bearer TOKEN_STRING"},
			expected: "TOKEN_STRING",
		},
		{
			input:    []string{"Authorization", ""},
			expected: "",
		},
		{
			input:    []string{"", "bearer TOKEN_STRING"},
			expected: "",
		},
	}
	for _, c := range cases {
		var header http.Header
		header.Set(c.input[0], c.input[1])
		token, err := GetBearerToken(header)
		if err != nil {
			t.Errorf("unable to retrieve token")
		}

		if token != c.expected {
			t.Errorf("token value: \"%s\" does not match expected value: \"%s\"", token, c.expected)
			continue
		}
	}
}
