# Azure Identity Terminator
This project aims to make it easier to leverage the [aad-pod-identity]() project's `Custom Resource Defintion (CRD)` in a Kubernetes cluster at scale when using individual pod identities. The easiest way to leverage a `Managed Identity` in an `Azure Kubernetes Service` cluster is to use the cluster's `Managed Identity`. The caveat to this is that every pod within the cluster now shares the same identity which is likely not going to satisfy most regulatory, compliance, or best securty practices.

However, the Operator for `aad-pod-identities` allows for leveraging individual `Serivce Principals` but this gets really difficult to manage at scale as cluster administrators attempt to keep up with all of the resources required for invidiual `aad-pod-identities`

**Azure Identity Terminator** attempts to solve the management overhead issues by using a `CRD` and an `Operator` within the cluster that will be able to:
- Create an Azure Active Directory Application Registration
- Generate a Service Principal and a random Client Secret
- Provide the required role assignment for the generated Service Principal
- Store the Client Secret in a Kubernetes Secret to be referenced by the AzureIdentity
- Create the AzureIdentity that leverages the new Serivce Principal and the aformentioned Kubernetes secret
- Finally, bind the identity using an AzureIdentityBinding which binds the AzureIdentity to the pod with its matching label

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

# Requirements
- Helm 3
- An Azure Kubernetes Service cluster running Kubernetes v1.16+
- kubectl v1.16+
- azure-cli 
- A general understanding of `aad-pod-identities`
- Access to your Azure Active Directory tenant to create the required Application Registration and Service Principal that AzureIdentityTerminator will leverage

# Installation
The first thing we need to do is generate a `Service Principal`
```az
az ad sp create-for-rbac --name azure-identity-terminator
```
