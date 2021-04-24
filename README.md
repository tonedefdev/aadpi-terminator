# Features
This project aims to make it easier to leverage the [aad-pod-identity](https://azure.github.io/aad-pod-identity) project's `Custom Resource Defintion (CRD)` in a Kubernetes cluster at scale when using individual pod identities. 

One of the easiest ways to leverage a `Managed Identity` in an `Azure Kubernetes Service` cluster is to use the VM/VMSS node's `Managed Identity`. The caveat to this is that every pod within the cluster now shares the same identity which is likely not going to satisfy most regulatory, compliance, or best securty practices. However, while the Operator for `aad-pod-identities` allows for leveraging individual `Serivce Principals` at scale this becomes an administrative burden.

**Azure Identity Terminator** attempts to solve the management overhead issues by using a `CRD` and an `Operator` within the cluster that will be able to:
- Create an Azure Active Directory Application Registration
- Generate a Service Principal and a random Client Secret
- Provide the required role assignment for the generated Service Principal
- Store the Client Secret in a Kubernetes Secret to be referenced by the AzureIdentity
- Create the AzureIdentity that leverages the new Serivce Principal and the aformentioned Kubernetes secret
- Bind the identity using an AzureIdentityBinding which binds the AzureIdentity to the pod with its matching label

The **Azure Identity Terminator System** is able to accomplish all of this by simply deploying an `AzureIdentityTerminator` manifest in the cluster as shown in the following example:
```yaml
apiVersion: azidterminator.io/v1alpha1
kind: AzureIdentityTerminator
metadata:
  name: azure-kv-access-test
  namespace: my-namespace
spec:
  appRegistration:
    displayName: azure-kv-access-test
  azureIdentityName: azure-kv-access-test
  nodeResourceGroup: my-aks-cluster-node-resource-group
  podSelector: azure-kv-pods
  servicePrincipal:
    clientSecretDuration: 720h
    tags:
    - azure-kv-aks-test
```
By abstracting away all of the steps required to create the necessary assets developers can simply declare the desired state for a pod's identity, and take the burden away from cluster operators who would inevitably have to manage these resources as the cluster scales.

Additionally, by adopting a `GitOps` workflow you can move your pod identity auditing to source control systems to have a full trail of the "*who/what/where/when*" these identities have been created.

Below is a guide that will walk you through setting up the **Azure Identity Terminator System** in your cluster.

# Getting Started
- Helm 3
- An Azure Kubernetes Service cluster running Kubernetes v1.16+
- kubectl v1.16+
- azure-cli 
- `aad-pod-identities` configured via this method: [AAD Pod Identity with Dedicated SP](https://azure.github.io/aad-pod-identity/docs/configure/deploy_aad_pod_dedicated_sp/)
- Access to your Azure Active Directory tenant to create the required Application Registration and Service Principal that AzureIdentityTerminator will leverage

# Pre-Requisites
The first thing we need to do is generate a `Service Principal`
```bash
az ad sp create-for-rbac --name azure-identity-terminator
```

Copy the output from this command and save it for the `values.yaml` file that we'll later pass into the Helm chart:
```yaml
# Default values for azure-identity-terminator.
# This is a YAML-formatted file.
# Declare variables to be passed into your templates.
replicaCount: 2
namespace: azid-terminator-system
deployment: azid-terminator-controller

secrets:
  azureClientID: <APP_ID>
  azureClientSecret: <PASSWORD>
  azureTenantID: <TENANT>
```

Save this file as `values.yaml`

Now, we need to give the `Service Principal` the necessary Graph API permissions to assign role assignments to the `AzureIdentity` objects that `AzureIdentityTeriminator` creates in addition to being able to create `Service Principals` to be used by the `AzureIdentity`

Run the following commands which will grant the `Service Principal` the permissions `Application.ReadWrite.OwnedBy` and `AppRoleAssignment.ReadWrite.All`
```bash
az ad app permission add --id <APP_ID> --api 00000002-0000-0000-c000-000000000000 --api-permissions 824c81eb-e3f8-4ee6-8f6d-de7f50d565b7=Role
az ad app permission add --id <APP_ID> --api 00000003-0000-0000-c000-000000000000 --api-permissions 06b708a9-e830-4db3-a914-8e69da51d44f=Role
```

>Follow the on-screen prompts for `az id permission grant` in order for these changes to take effect. 

`Application.ReadWrite.OwnedBy` is required to allow the `Service Principal` to create and manage only the `Service Principals` that `AzureIdentityTerminator` creates. This ensures that the `Service Principal` is not able to modify any other `App Registrations` in the Azure AD tenant besides the ones that are created by `AzureIdentityTerminator`

`AppRoleAssignment.ReadWrite.All` is reqiured to give the `Service Princiapls` that are created by `AzureIdentityTerminator` the `Reader` role over the node resource group where the AKS cluster scaleset resides. This is required in order to allow the `ServicePrincipal` to be bound to the underlying node otherwise you would not be able to leverage the `AzureIdentity` with a pod.

Run the following commands to grant the required admin consent to allow these permissions to be consented:
```bash
az ad app permission admin-consent --id <APP_ID>
```

# Installation
We'll use Helm to deploy the application to the cluster. First add the application's repo:
```bash
helm repo add azure-identity-terminator https://tonedefdev.github.io/azure-identity-terminator/
```

Then we'll need to update the repo:
```bash
helm repo update
```

Now we can install the application into the cluster:
```bash
helm install azure-identity-terminator azure-identity-terminator/azure-identity-terminator --create-namespace --namespace azure-identity-terminator -f values.yaml
```

Once successfully installed you can check the pods are running:
```bash
kubectl get pods -n azure-identity-terminator
```

# Deploy a Terminator
First we need to create an `AzureIdentityManfiest`:
```yaml
apiVersion: azidterminator.io/v1alpha1
kind: AzureIdentityTerminator
metadata:
  name: azure-kv-access-test
  namespace: my-namespace
spec:
  appRegistration:
    displayName: azure-kv-access-test
  azureIdentityName: azure-kv-access-test
  nodeResourceGroup: my-aks-cluster-node-resource-group
  podSelector: azure-kv-pods
  servicePrincipal:
    clientSecretDuration: 720h
    tags:
    - azure-kv-aks-test
```

The fields of this definition should be pretty self-explanatory. You'll need to supply all fields with the `tags` being optional. The `tags` help for automation purposes where you may want to automate the rotation of `AzureIdentities` as they expire, so setting appropriate tags can help you find and locate the service principals that require rotation.

Once we have saved our manifest we can apply it to the cluster:
```bash
kubectl apply -f azidterminator.yaml
```

Once the manifest has been deployed the `AzureIdentityTerminator` controller will pick up the manifest and start processing it. You can tail the logs of the active controller pod by running:
```bash
kubectl logs -n azure-identity-terminator <active_pod> -f
```

You should see a number of success messages that indicate the `AzureIdentityTerminator` has successfully generated all the resources required for the `aad-pod-identity` controller to bind the `Service Principal` to the cluster node.

You can view the actual `AzureIdentityTerminator` status by running:
```bash
kubectl get azid -n my-namespace
```

