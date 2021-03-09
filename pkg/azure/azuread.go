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

// CreateApp interface contains methods to create Azure AD Application resources
type CreateApp interface {
	CreateAzureADApp() graphrbac.Application
	CreateServicePrincipal() graphrbac.ServicePrincipal
}

// App struct defines an Azure AD Application
type App struct {
	ClientID     string
	ClientSecret string
	DisplayName  string
	Duration     int64
	TenantID     string
}

func newAuthorizer() autorest.Authorizer {
	authorizer, err := auth.NewAuthorizerFromEnvironment()
	if err != nil {
		return nil
	}
	return authorizer
}

func generateRandomSecret() string {
	randomPassword := uuid.New()
	return randomPassword.String()
}

// CreateAzureADApp creates an Azure AD Application
func (aadapp *App) CreateAzureADApp() graphrbac.Application {
	appClient := graphrbac.NewApplicationsClient(os.Getenv("AZURE_TENANT_ID"))
	appClient.Authorizer = newAuthorizer()

	appCreateParam := graphrbac.ApplicationCreateParameters{
		DisplayName:             to.StringPtr(aadapp.DisplayName),
		AvailableToOtherTenants: to.BoolPtr(false),
	}

	appReg, err := appClient.Create(context.Background(), appCreateParam)
	if err != nil {
		// TODO: Implement logger
	}

	aadapp.TenantID = os.Getenv("AZURE_TENANT_ID")
	aadapp.ClientID = *appReg.AppID
	return appReg
}

// CreateServicePrincipal generates a service princiapl for an AzureIdentityTerminator resource
func (aadapp *App) CreateServicePrincipal() graphrbac.ServicePrincipal {
	spnClient := graphrbac.NewServicePrincipalsClient(os.Getenv("AZURE_TENANT_ID"))
	spnClient.Authorizer = NewAuthorizer()
	secret := generateRandomSecret()

	var now *date.Time
	var expiration *date.Time

	*now = date.Time{time.Now()}
	*expiration = date.Time{
		time.Now().Add(time.Hour * time.Duration(aadapp.Duration)),
	}

	var clientSecret *[]graphrbac.PasswordCredential

	newClientSecret := graphrbac.PasswordCredential{
		StartDate: now,
		EndDate:   expiration,
		Value:     to.StringPtr(secret),
	}

	*clientSecret = append(*clientSecret, newClientSecret)

	spnCreateParam := graphrbac.ServicePrincipalCreateParameters{
		AppID:               to.StringPtr(aadapp.ClientID),
		PasswordCredentials: clientSecret,
	}

	spnCreate, err := spnClient.Create(context.Background(), spnCreateParam)
	if err != nil {
		// TODO: Implement logger
	}

	aadapp.ClientSecret = secret
	return spnCreate
}
