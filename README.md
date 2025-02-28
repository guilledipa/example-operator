# Example Operator

This is an initial attempt to build a K8s operator.

## Steps

1. Install kubebuilder

    ```shell
    curl -L -o kubebuilder "https://go.kubebuilder.io/dl/latest/$(go env GOOS)/$(go env GOARCH)"
    chmod +x kubebuilder && sudo mv kubebuilder /usr/local/bin/
    ```

1. Create a new project directory

    ```shell
    mkdir example-operator
    cd example-operator
    ```

1. Initialize a new operator project

    ```shell
    kubebuilder init --domain example.com --repo github.com/guilledipa/example-operator
    ```

1. Create an API and controller

    ```shell
    kubebuilder create api --group apps --version v1alpha1 --kind Example
    ```

1. Define the CR in `api/v1alpha1/example_types.go`

1. Implemet the controller in `controllers/example_controller.go`

1. Generate manifests

    ```shell
    make manifests
    ```

1. Install CRDs into the cluster

    ```shell
    make install
    ```

1. Build and push your operator image (if using a remote cluster)

    ```shell
    make docker-build docker-push IMG=example-operator:v0.1.0
    ```

1. Deploy the operator

    ```shell
    make deploy IMG=localhost/example-operator:v0.1.0
    ```
