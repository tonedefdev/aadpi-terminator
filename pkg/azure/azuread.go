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
	"github.com/go-logr/logr"
	"github.com/google/uuid"
)

// CreateApp interface contains methods to create Azure AD Application resources
type CreateApp interface {
	CreateAzureADApp() graphrbac.Application
	CreateServicePrincipal() graphrbac.ServicePrincipal
	Logger() logr.Logger
}

// App struct defines an Azure AD Application
type App struct {
	ClientID     string
	ClientSecret string
	DisplayName  string
	Duration     int64
	Log          logr.Logger
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

// Logger creates a logging instance
func (aadApp *App) Logger() logr.Logger {
	log := aadApp.Log.WithValues("azureidentityterminator", "appregistration")
	return log
}

// CreateAzureADApp creates an Azure AD Application
func (aadApp *App) CreateAzureADApp() graphrbac.Application {
	log := aadApp.Logger()
	appClient := graphrbac.NewApplicationsClient(os.Getenv("AZURE_TENANT_ID"))
	appClient.Authorizer = newAuthorizer()

	appCreateParam := graphrbac.ApplicationCreateParameters{
		DisplayName:             to.StringPtr(aadApp.DisplayName),
		AvailableToOtherTenants: to.BoolPtr(false),
	}

	appReg, err := appClient.Create(context.Background(), appCreateParam)
	if err != nil {
		log.Error(err, "AzureIdentityTerminator Azure AD App Registration Failed", appCreateParam.DisplayName)
	}

	aadApp.TenantID = os.Getenv("AZURE_TENANT_ID")
	aadApp.ClientID = *appReg.AppID
	return appReg
}

// CreateServicePrincipal generates a service princiapl for an AzureIdentityTerminator resource
func (aadApp *App) CreateServicePrincipal() graphrbac.ServicePrincipal {
	log := aadApp.Logger()
	spnClient := graphrbac.NewServicePrincipalsClient(os.Getenv("AZURE_TENANT_ID"))
	spnClient.Authorizer = newAuthorizer()
	secret := generateRandomSecret()

	var now *date.Time
	var expiration *date.Time

	*now = date.Time{time.Now()}
	*expiration = date.Time{
		time.Now().Add(time.Hour * time.Duration(aadApp.Duration)),
	}

	var clientSecret *[]graphrbac.PasswordCredential

	newClientSecret := graphrbac.PasswordCredential{
		StartDate: now,
		EndDate:   expiration,
		Value:     to.StringPtr(secret),
	}

	*clientSecret = append(*clientSecret, newClientSecret)

	spnCreateParam := graphrbac.ServicePrincipalCreateParameters{
		AppID:               to.StringPtr(aadApp.ClientID),
		PasswordCredentials: clientSecret,
	}

	spnCreate, err := spnClient.Create(context.Background(), spnCreateParam)
	if err != nil {
		log.Error(err, "AzureIdentityTerminator Service Principal Creation failed", spnCreateParam.AppID)
	}

	aadApp.ClientSecret = secret
	return spnCreate
}
