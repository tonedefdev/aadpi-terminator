package azuread

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/authorization/mgmt/2015-07-01/authorization"
	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/google/uuid"
	iam "github.com/tonedefdev/aadpi-terminator/pkg/iam"
	config "github.com/tonedefdev/aadpi-terminator/pkg/internal"
)

// App struct defines an Azure AD Application
type App struct {
	ClientID               string
	ClientSecret           string
	ClientSecretExpiration date.Time
	DisplayName            string
	Duration               string
	NodeResourceGroupID    string
	ObjectID               string
	ServicePrincipalID     string
	ServicePrincipalTags   []string
	TenantID               string
}

// Adds the provided SPN to the 'Reader' role for the AKS cluster node resource group
func createRoleAssignment(aadApp *App) error {
	ctx := context.Background()
	sub := config.SubscriptionID()
	reader := "/subscriptions/" + sub + "/providers/Microsoft.Authorization/roleDefinitions/acdd72a7-3385-48ef-bd42-f606fba81ae7"

	roleAssignmentsClient, _ := getRoleAssignmentsClient()
	_, err := roleAssignmentsClient.Create(
		ctx,
		aadApp.NodeResourceGroupID,
		uuid.New().String(),
		authorization.RoleAssignmentCreateParameters{
			Properties: &authorization.RoleAssignmentProperties{
				PrincipalID:      to.StringPtr(aadApp.ServicePrincipalID),
				RoleDefinitionID: to.StringPtr(reader),
			},
		})

	return err
}

func generateRandomSecret() string {
	randomPassword := uuid.New()
	return randomPassword.String()
}

func getApplicationsClient() graphrbac.ApplicationsClient {
	appClient := graphrbac.NewApplicationsClient(config.TenantID())
	a, _ := iam.GetGraphAuthorizer()
	appClient.Authorizer = a
	appClient.AddToUserAgent(config.UserAgent())
	return appClient
}

func getRoleAssignmentsClient() (authorization.RoleAssignmentsClient, error) {
	roleClient := authorization.NewRoleAssignmentsClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	roleClient.Authorizer = a
	roleClient.AddToUserAgent(config.UserAgent())
	return roleClient, nil
}

func getServicePrincipalClient() graphrbac.ServicePrincipalsClient {
	spnClient := graphrbac.NewServicePrincipalsClient(config.TenantID())
	a, _ := iam.GetGraphAuthorizer()
	spnClient.Authorizer = a
	spnClient.AddToUserAgent(config.UserAgent())
	return spnClient
}

// CreateAzureADApp creates an Azure AD Application
func (aadApp *App) CreateAzureADApp() (graphrbac.Application, error) {
	ctx := context.Background()
	appClient := getApplicationsClient()

	appCreateParam := graphrbac.ApplicationCreateParameters{
		DisplayName:             to.StringPtr(aadApp.DisplayName),
		AvailableToOtherTenants: to.BoolPtr(false),
	}

	appReg, err := appClient.Create(ctx, appCreateParam)
	if err != nil {
		return appReg, err
	}

	aadApp.ClientID = *appReg.AppID
	aadApp.ObjectID = *appReg.ObjectID
	aadApp.TenantID = config.TenantID()
	return appReg, err
}

// CreateServicePrincipal generates a service princiapl for an AzureIdentityTerminator resource
func (aadApp *App) CreateServicePrincipal() (graphrbac.ServicePrincipal, error) {
	ctx := context.Background()
	spnClient := getServicePrincipalClient()
	secret := generateRandomSecret()

	duration, err := time.ParseDuration(aadApp.Duration)
	if err != nil {
		// TODO: Implement error handling
	}

	now := &date.Time{
		Time: time.Now(),
	}
	expiration := &date.Time{
		Time: time.Now().Add(duration),
	}

	var clientSecret = []graphrbac.PasswordCredential{}

	newClientSecret := graphrbac.PasswordCredential{
		StartDate: now,
		EndDate:   expiration,
		Value:     to.StringPtr(secret),
	}

	clientSecret = append(clientSecret, newClientSecret)

	spnCreateParam := graphrbac.ServicePrincipalCreateParameters{
		AppID:               to.StringPtr(aadApp.ClientID),
		PasswordCredentials: &clientSecret,
		Tags:                &aadApp.ServicePrincipalTags,
	}

	spnCreate, err := spnClient.Create(ctx, spnCreateParam)
	if err != nil {
		return spnCreate, err
	}

	aadApp.ClientSecret = secret
	aadApp.ClientSecretExpiration = *expiration
	aadApp.ServicePrincipalID = *spnCreate.ObjectID

	for {
		err = createRoleAssignment(aadApp)
		if err == nil {
			break
		} else {
			fmt.Println(err)
		}
	}

	return spnCreate, err
}

// DeleteAzureApp deletes the requested Azure AD application
func (aadApp *App) DeleteAzureApp() (autorest.Response, error) {
	ctx := context.Background()
	appClient := getApplicationsClient()

	appDelete, err := appClient.Delete(ctx, aadApp.ObjectID)
	if err != nil {
		return appDelete, err
	}

	return appDelete, err
}
