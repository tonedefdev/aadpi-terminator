package iam

import (
	"fmt"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	config "github.com/tonedefdev/aadpi-terminator/pkg/internal"
)

var (
	graphAuthorizer autorest.Authorizer
)

// OAuthGrantType specifies which grant type to use.
type OAuthGrantType int

const (
	// OAuthGrantTypeServicePrincipal for client credentials flow
	OAuthGrantTypeServicePrincipal OAuthGrantType = iota
	// OAuthGrantTypeDeviceFlow for device flow
	OAuthGrantTypeDeviceFlow
)

// GrantType returns what grant type has been configured.
func grantType() OAuthGrantType {
	if config.UseDeviceFlow() {
		return OAuthGrantTypeDeviceFlow
	}
	return OAuthGrantTypeServicePrincipal
}

// GetGraphAuthorizer gets an OAuthTokenAuthorizer for graphrbac API.
func GetGraphAuthorizer() (autorest.Authorizer, error) {
	if graphAuthorizer != nil {
		return graphAuthorizer, nil
	}

	var a autorest.Authorizer
	var err error

	a, err = getAuthorizerForResource(grantType(), config.Environment().GraphEndpoint)

	if err == nil {
		// cache
		graphAuthorizer = a
	} else {
		graphAuthorizer = nil
	}

	return graphAuthorizer, err
}

func getAuthorizerForResource(grantType OAuthGrantType, resource string) (autorest.Authorizer, error) {
	var a autorest.Authorizer
	var err error

	switch grantType {

	case OAuthGrantTypeServicePrincipal:
		oauthConfig, err := adal.NewOAuthConfig(
			config.Environment().ActiveDirectoryEndpoint, config.TenantID())
		if err != nil {
			return nil, err
		}

		token, err := adal.NewServicePrincipalToken(
			*oauthConfig, config.ClientID(), config.ClientSecret(), resource)
		if err != nil {
			return nil, err
		}
		a = autorest.NewBearerAuthorizer(token)

	case OAuthGrantTypeDeviceFlow:
		deviceconfig := auth.NewDeviceFlowConfig(config.ClientID(), config.TenantID())
		deviceconfig.Resource = resource
		a, err = deviceconfig.Authorizer()
		if err != nil {
			return nil, err
		}

	default:
		return a, fmt.Errorf("invalid grant type specified")
	}

	return a, err
}
