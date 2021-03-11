/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	aadpodv1 "github.com/Azure/aad-pod-identity/pkg/apis/aadpodidentity/v1"
	"github.com/Azure/go-autorest/autorest/to"

	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	terminatorv1alpha1 "github.com/tonedefdev/aadpi-terminator/api/v1alpha1"
	azuread "github.com/tonedefdev/aadpi-terminator/pkg/azure"
)

// AzureIdentityTerminatorReconciler reconciles a AzureIdentityTerminator object
type AzureIdentityTerminatorReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=aadpi-terminator.io,resources=azureidentityterminators,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=aadpi-terminator.io,resources=azureidentityterminators/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=aadpi-terminator.io,resources=azureidentityterminators/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=aadpodidentity.k8s.io,resources=azureidentities;azureidentitybindings,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *AzureIdentityTerminatorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("azureidentityterminator", req.NamespacedName)
	ctx = context.Background()

	// Fetch the AzureIdentityTerminator instance
	terminator := &terminatorv1alpha1.AzureIdentityTerminator{}
	err := r.Get(ctx, req.NamespacedName, terminator)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("AzureIdentityTerminator resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get AzureIdentityTerminator")
		return ctrl.Result{}, err
	}

	aadID := &aadpodv1.AzureIdentity{}
	err = r.Get(ctx, types.NamespacedName{Name: terminator.Name, Namespace: terminator.Namespace}, aadID)
	if err != nil && errors.IsNotFound(err) {

		// Create the Azure AD Application that the AzureIdentity will leverage
		aadAppRegistration := r.CreateApp(terminator.Spec.AADRegistrationName, terminator.Spec.ClientSecretDuration)
		log.Info("Creating a new Azure AD App Registration", "ClientID", aadAppRegistration.ClientID, "Display Name", aadAppRegistration.DisplayName)

		secret := &corev1.Secret{}
		err = r.Get(ctx, types.NamespacedName{Name: terminator.Name, Namespace: terminator.Namespace}, secret)
		if err != nil && errors.IsNotFound(err) {

			// Create Secret that will contain the ClientSecret for the AzureIdentity
			sec := r.SecretManfiest(terminator, aadAppRegistration)
			err = r.Create(ctx, sec)
			if err != nil {
				log.Error(err, "Failed to create new Secret", "Secret.Namespace", sec.Namespace, "Secret.Name", sec.Name)
				return ctrl.Result{}, err
			}

			// Create AzureIdentity
			azID := r.AzureIdentityManifest(terminator, aadAppRegistration)
			err = r.Create(ctx, azID)
			if err != nil {
				log.Error(err, "Failed to create AzureIdentity", "AzureIdentity.Name", azID.Name)
				return ctrl.Result{}, err
			}

			// Create AzureIdentityBinding
			azIDBinding := r.AzureIdentityBindingManifest(terminator, azID)
			err = r.Create(ctx, azIDBinding)
			if err != nil {
				log.Error(err, "Failed to create AzureIdentityBinding", "AzureIdentityBinding.Name", azIDBinding.Name)
			}
		}

		// AzureIdentity and AzureIdentityBinding created with new Azure AD App successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil

	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// CreateApp creates the Azure AD Application, SPN, and returns the necessary information
func (r *AzureIdentityTerminatorReconciler) CreateApp(aadRegName string, duration int64) *azuread.App {
	app := &azuread.App{
		DisplayName: aadRegName,
		Duration:    duration,
	}
	app.CreateAzureADApp()
	app.CreateServicePrincipal()
	return app
}

// AzureIdentityManifest creates the AzureIdentity manifest
func (r *AzureIdentityTerminatorReconciler) AzureIdentityManifest(t *terminatorv1alpha1.AzureIdentityTerminator, app *azuread.App) *aadpodv1.AzureIdentity {
	azID := &aadpodv1.AzureIdentity{
		TypeMeta: v1.TypeMeta{
			Kind:       "AzureIdentity",
			APIVersion: "aadpodidentity.k8s.io",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      t.Name,
			Namespace: t.Namespace,
		},
		Spec: aadpodv1.AzureIdentitySpec{
			Type:     aadpodv1.IdentityType(1),
			ClientID: app.ClientID,
			ClientPassword: corev1.SecretReference{
				Name:      t.Name,
				Namespace: t.Namespace,
			},
		},
	}

	return azID
}

// AzureIdentityBindingManifest creates teh AzureIdentityBinding manifest
func (r *AzureIdentityTerminatorReconciler) AzureIdentityBindingManifest(t *terminatorv1alpha1.AzureIdentityTerminator, azID *aadpodv1.AzureIdentity) *aadpodv1.AzureIdentityBinding {
	azIDBinding := &aadpodv1.AzureIdentityBinding{
		TypeMeta: v1.TypeMeta{
			Kind:       "AzureIdentityBinding",
			APIVersion: "aadpodidentity.k8s.io",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      t.Name,
			Namespace: t.Namespace,
		},
		Spec: aadpodv1.AzureIdentityBindingSpec{
			AzureIdentity: azID.Name,
			Selector:      t.Spec.PodSelector,
		},
	}

	return azIDBinding
}

// SecretManfiest creates the Secret needed for the AzureIdentity
func (r *AzureIdentityTerminatorReconciler) SecretManfiest(t *terminatorv1alpha1.AzureIdentityTerminator, app *azuread.App) *corev1.Secret {
	secret := &corev1.Secret{
		TypeMeta: v1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      t.Name,
			Namespace: t.Namespace,
		},
		Immutable: to.BoolPtr(true),
		StringData: map[string]string{
			app.ClientID: app.ClientSecret,
		},
	}

	return secret
}

// SetupWithManager sets up the reconciler management
func (r *AzureIdentityTerminatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&terminatorv1alpha1.AzureIdentityTerminator{}).
		Complete(r)
}
