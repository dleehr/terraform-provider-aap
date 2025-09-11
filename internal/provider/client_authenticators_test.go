package provider

import (
	"net/http"
	"strings"
	"testing"
)

func TestNewBasicAuthenticator(t *testing.T) {
	testUsername := "testusername"
	testPassword := "testpassword"
	var testTable = []struct {
		name          string
		username      *string
		password      *string
		expectSuccess bool
	}{
		{
			name:          "Success when providing username and password",
			username:      &testUsername,
			password:      &testPassword,
			expectSuccess: true,
		},
		{
			name:          "Failure when username and password are nil",
			username:      nil,
			password:      nil,
			expectSuccess: false,
		},
		{
			name:          "Failure when username is provided and password is nil",
			username:      &testUsername,
			password:      nil,
			expectSuccess: false,
		},
		{
			name:          "Failure with when username is nil and password is provided",
			username:      nil,
			password:      &testPassword,
			expectSuccess: false,
		},
	}
	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			auth, diags := NewBasicAuthenticator(test.username, test.password)
			if test.expectSuccess {
				if auth == nil {
					t.Errorf("Expected NewBasicAuthenticator result to be defined, failed with %v", diags)
				}
			} else {
				if !diags.HasError() {
					t.Errorf("Expected NewBasicAuthenticator to fail, received %v", auth)
				}
			}
		})
	}
}

func TestBasicAuthenticatorConfigure(t *testing.T) {
	testUsername := "username"
	testPassword := "password"
	auth, _ := NewBasicAuthenticator(&testUsername, &testPassword)
	req, _ := http.NewRequest("", "", strings.NewReader(""))
	auth.Configure(req)
	actual := req.Header["Authorization"][0]
	expected := "Basic dXNlcm5hbWU6cGFzc3dvcmQ=" // username:password in base64
	if actual != expected {
		t.Errorf("Expected (%s) not equal to actual (%s)", expected, actual)
	}
}

func TestNewTokenAuthenticator(t *testing.T) {
	testToken := "testtoken"
	testHeader := "testheader"
	testPrefix := "testprefix"
	var testTable = []struct {
		name          string
		token         *string
		prefix        *string
		header        *string
		expectSuccess bool
	}{
		{
			name:          "Success when token is provided and prefix/header are nil",
			token:         &testToken,
			prefix:        nil,
			header:        nil,
			expectSuccess: true,
		},
		{
			name:          "Success when token, prefix, and header are provided",
			token:         &testToken,
			prefix:        &testPrefix,
			header:        &testHeader,
			expectSuccess: true,
		},
		{
			name:          "Failure when token is nil and prefix/header are provided",
			token:         nil,
			prefix:        &testPrefix,
			header:        &testHeader,
			expectSuccess: false,
		},
		{
			name:          "Failure when token, prefix, and header are nil",
			token:         nil,
			prefix:        nil,
			header:        nil,
			expectSuccess: false,
		},
	}
	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			auth, diags := NewTokenAuthenticator(test.token, test.prefix, test.header)
			if test.expectSuccess {
				if auth == nil {
					t.Errorf("Expected NewTokenAuthenticator result to be defined, failed with %v", diags)
				}
			} else {
				if !diags.HasError() {
					t.Errorf("Expected NewTokenAuthenticator to fail, received %v", auth)
				}
			}
		})
	}
}

func TestTokenAuthenticatorConfigure(t *testing.T) {
	testToken := "testtoken"
	testPrefix := "Token"
	testHeader := "Authorizationtest"

	var testTable = []struct {
		name         string
		token        *string
		prefix       *string
		header       *string
		expectHeader string
		expectValue  string
	}{
		{
			name:         "Configure defaults header to Authorization: Bearer ...",
			token:        &testToken,
			prefix:       nil,
			header:       nil,
			expectHeader: "Authorization",
			expectValue:  "Bearer testtoken",
		},
		{
			name:         "Configure honors prefix and header",
			token:        &testToken,
			prefix:       &testPrefix,
			header:       &testHeader,
			expectHeader: "Authorizationtest",
			expectValue:  "Token testtoken",
		},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			auth, _ := NewTokenAuthenticator(test.token, test.prefix, test.header)
			req, _ := http.NewRequest("", "", strings.NewReader(""))
			auth.Configure(req)
			actual := req.Header[test.expectHeader][0]
			expected := test.expectValue
			if actual != expected {
				t.Errorf("Expected (%s) not equal to actual (%s)", expected, actual)
			}
		})
	}
}
