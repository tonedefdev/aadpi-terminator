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

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/tonedefdev/aadpi-terminator/pkg/azure/azuread"

	terminatorv1alpha1 "github.com/tonedefdev/aadpi-terminator/api/v1alpha1"
)

// AzureIdentityTerminatorReconciler reconciles a AzureIdentityTerminator object
type AzureIdentityTerminatorReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=aadpi-terminator.k8s.io,resources=azureidentityterminators,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=aadpi-terminator.k8s.io,resources=azureidentityterminators/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=aadpi-terminator.k8s.io,resources=azureidentityterminators/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *AzureIdentityTerminatorReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("azureidentityterminator", req.NamespacedName)
	ctx := context.Background()

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

	aadpod := &aadpodv1.AzureIdentity{}
	err = r.Get(ctx, types.NamespacedName{Name: terminator.Name, Namespace: terminator.Namespace}, aadpod)
	if err != nil && errors.IsNotFound(err) {
		//
		aadAppRegistration := r.CreateApp(terminator.Spec.AADRegistrationName)
		log.Info("Creating a new Azure AD App Registration", "AzureIdentityTerminator.Namespace", dep.Namespace, "AzureIdentityTerminator.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *AzureIdentityTerminatorReconciler) CreateApp(aadRegName string, duration int64) {
	app := &azuread.CreateAzureADApp(aadRegName)

}

func (r *AzureIdentityTerminatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&terminatorv1alpha1.AzureIdentityTerminator{}).
		Owns(&aadpodv1.AzureIdentity{}).
		Complete(r)
}
