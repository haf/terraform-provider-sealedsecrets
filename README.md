# Terraform Provider: `haf/sealedsecrets`

This is a fork of [rockyhmchen's terraform-provider-sealedsecrets](https://github.com/rockyhmchen/terraform-provider-sealedsecrets), published under [kita99/sealedsecrets](https://registry.terraform.io/providers/kita99/sealedsecrets/latest).

This is a fork of [kita99's changes](https://github.com/haf/terraform-provider-sealedsecrets) to vet the code, and
document it, aiming to send a PR back.

The `sealedsecrets` provider helps you manage SealedSecret objects (`bitnami.com/v1alpha1`) from terraform. It generates a
K8s Secret from the key/value pairs you give as input, encrypts it using `kubeseal` and finally applies it to the cluster.
In subsequent runs it will check if the object still exists in the cluster or if the contents have changed and act accordingly.

[How to test](https://faun.pub/lets-do-devops-compile-and-test-local-terraform-provider-6f056b69c587)


### Usage

```HCL
terraform {
  required_providers {
    sealedsecrets = {
      source = "haf/sealedsecrets"
      version = "0.2.4"
    }
  }
}

data "google_container_cluster" "main" {
  name     = "main"
  project  = "logary-prod"
  location = "europe-west4"
}

provider "kubernetes" {
  host                   = "https://${data.google_container_cluster.main.endpoint}"
  token                  = data.google_client_config.provider.access_token
  cluster_ca_certificate = base64decode(data.google_container_cluster.main.master_auth[0].cluster_ca_certificate)
}

provider "sealedsecrets" {
  kubernetes = provider.kubernetes
}

resource "sealedsecrets_secret" "pg_credentials" {
  type                 = "Opaque"
  name                 = "pg-credentials"
  namespace            = "apps"
  controller_name      = "sealed-secret-controller"
  controller_namespace = "kube-system"
  depends_on           = []
  secrets = {
    key = "value"
  }
  annotations = {
    key = "value"
  }
  labels = {
    key = "value"
  }
}
```

### Argument Reference

The following arguments are supported:

- `type` -  The secret type. ex: `Opaque`
- `name` - Name of the secret, must be unique.
- `namespace` - Namespace defines the space within which name of the secret must be unique.
- `controller_name` - Name of the SealedSecrets controller in the cluster
- `controller_namespace` - Namespace of the SealedSecrets controller in the cluster
- `depends_on` - For specifying hidden dependencies.
- `secrets` - Key/value pairs to populate the secret (can be empty, but not really useful here)
- `annotations` - Key/value pairs to populate the secret (can be empty)
- `labels` - Key/value pairs to populate the secret (can be empty)

*NOTE: All the arguments above are required*


### Behind the scenes

#### Create

Takes resource inputs to form the below command, computes SHA256 hash of the resulting SealedSecret manifest and sets it as the resource id.

```bash
kubectl create secret generic {sealedsecrets_secret.[resource].name} \
  --namespace={sealedsecrets_secret.[resource].namespace} \
  --type={sealedsecrets_secret.[resource].type} \
  --from-literal={sealedsecrets_secret.[resource].secrets.key}={sealedsecrets_secret.[resource].secrets.value} \ # line repeated for each key/value pair
  --dry-run \
  --output=yaml | \
  kubeseal \    
    --controller-name ${sealedsecrets_secret.[resource].controller_name} \
    --controller-namespace ${sealedsecrets_secret.[resource].controller_namespace} \
    --format yaml \
    > /tmp/sealedsecret.yaml
```


#### Read

Checks if the SealedSecret object still exists in the cluster or if the SHA256 hash has changed.


#### Update

Same as `Create`


#### Delete

Removes SealedSecret object from the cluster and deletes terraform state.


# References

- https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2@v2.6.1
- https://registry.terraform.io/publish/provider/github/haf
- https://registry.terraform.io/settings/gpg-keys
