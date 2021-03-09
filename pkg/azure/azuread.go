package azuread

import (
	"context"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/google/uuid"
)

type CreateAzureADApp interface {
	CreateAzureADApp() graphrbac.Application
	CreateServicePrincipal() graphrbac.ServicePrincipal
}

type AzureADApp struct {
	ClientID     string
	ClientSecret string
	DisplayName  string
}

// NewAuthorizer returns an authorizer using envrionment variables
func NewAuthorizer() autorest.Authorizer {
	authorizer, err := auth.NewAuthorizerFromEnvironment()
	if err != nil {
		return nil
	}
	return authorizer
}

// CreateAzureADApp creates an Azure AD Application
func (aadp *AzureADApp) CreateAzureADApp() graphrbac.Application {
	appClient := graphrbac.NewApplicationsClient(os.Getenv("AZURE_TENANT_ID"))
	appClient.Authorizer = NewAuthorizer()

	appCreateParam := graphrbac.ApplicationCreateParameters{
		DisplayName:             to.StringPtr(aadp.DisplayName),
		AvailableToOtherTenants: to.BoolPtr(false),
	}

	appReg, err := appClient.Create(context.Background(), appCreateParam)
	if err != nil {
		// TODO: Implement logger
	}

	*aadp.ClientID = *appReg.AppID
	return appReg
}

// GenerateRandomSecret creates a random UUID to be used as the client secret
func GenerateRandomSecret() string {
	randomPassword := uuid.New()
	return randomPassword.String()
}

// CreateServicePrincipal generates a service princiapl for an AzureIdentityTerminator resource
func (aadp *AzureADApp) CreateServicePrincipal(applicationID string, duration int64) graphrbac.ServicePrincipal {
	spnClient := graphrbac.NewServicePrincipalsClient(os.Getenv("AZURE_TENANT_ID"))
	spnClient.Authorizer = NewAuthorizer()
	secret := GenerateRandomSecret()

	var now *date.Time
	var expiration *date.Time

	*now = date.Time{time.Now()}
	*expiration = date.Time{
		time.Now().Add(time.Hour * time.Duration(duration)),
	}

	var clientSecret *[]graphrbac.PasswordCredential

	newClientSecret := graphrbac.PasswordCredential{
		StartDate: now,
		EndDate:   expiration,
		Value:     to.StringPtr(secret),
	}

	*clientSecret = append(*clientSecret, newClientSecret)

	spnCreateParam := graphrbac.ServicePrincipalCreateParameters{
		AppID:               to.StringPtr(applicationID),
		PasswordCredentials: clientSecret,
	}

	spnCreate, err := spnClient.Create(context.Background(), spnCreateParam)
	if err != nil {
		// TODO: Implement logger
	}

	*aadp.ClientSecret = secret
	return spnCreate
}
