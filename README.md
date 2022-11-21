# oci-privateca-issuer

** WORK IN PROGRESS **

cert-manager external issuer for oci certificates

## Description

cert-manager external issuer for oci certificates

### Install CRDs
Install Instances of Custom Resources:

```sh
make install
```

### Uninstall CRDs
To delete the CRDs from the cluster:

```sh
make uninstall
```

### Deploy controller
UnDeploy the controller to the cluster:

```sh
make deploy
```

### Undeploy controller
UnDeploy the controller to the cluster:

```sh
make undeploy
```

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/) 
which provides a reconcile function responsible for synchronizing resources untile the desired state is reached on the cluster 

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)


