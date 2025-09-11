package provider

import (
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type AAPClientAuthenticator interface {
	Configure(*http.Request)
}

// Basic authenticator supports username/password auth
type AAPClientBasicAuthenticator struct {
	username string
	password string
}

// Should this return a pointer to a struct or should it return an interface?
// "Accept interfaces, return structs"
func NewBasicAuthenticator(username *string, password *string) (*AAPClientBasicAuthenticator, diag.Diagnostics) {
	var diags diag.Diagnostics
	if username == nil {
		diags.AddError(
			"Missing username",
			"Unable to create a basic authenticator without username")
	}
	if password == nil {
		diags.AddError(
			"Missing password",
			"Unable to create a basic authenticator without password")
	}
	if diags.HasError() {
		return nil, diags
	}
	return &AAPClientBasicAuthenticator{
		username: *username,
		password: *password,
	}, nil
}

func (a *AAPClientBasicAuthenticator) Configure(req *http.Request) {
	// To configure basic auth, we can just use http.Request's SetBasicAuth
	req.SetBasicAuth(a.username, a.password)
}

// Token authenticator supports Token auth
type AAPClientTokenAuthenticator struct {
	token  string // Required
	prefix string // Optional, defaults to "Bearer"
	header string // Optional, defaults to "Authorization"
	// Do we need a refresh token?
}

func NewTokenAuthenticator(token *string, prefix *string, header *string) (*AAPClientTokenAuthenticator, diag.Diagnostics) {
	var diags diag.Diagnostics
	if token == nil {
		// token must be supplied. If not, that's an error
		diags.AddError(
			"Missing token",
			"Unable to create a token authenticator without token")
	}
	if header == nil {
		*header = "Authorization"
	}
	if prefix == nil {
		*prefix = "Bearer"
	}

	if diags.HasError() {
		return nil, diags
	}

	return &AAPClientTokenAuthenticator{
		token:  *token,
		prefix: *prefix,
		header: *header,
	}, nil
}

func (a *AAPClientTokenAuthenticator) Configure(req *http.Request) {
	req.Header.Set(a.header, fmt.Sprintf("%s %s", a.prefix, a.token))
}
