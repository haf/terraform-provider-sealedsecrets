terraform-provider-sealedsecrets
================================

This is a fork of [rockyhmchen's terraform-provider-sealedsecrets](https://github.com/rockyhmchen/terraform-provider-sealedsecrets), published under [kita99/sealedsecrets](https://registry.terraform.io/providers/kita99/sealedsecrets/latest).

This is a fork of [kita99's changes](https://github.com/haf/terraform-provider-sealedsecrets) to vet the code, and
document it, aiming to send a PR back.

The `sealedsecrets` provider helps you manage SealedSecret objects (`bitnami.com/v1alpha1`) from terraform. It generates a
K8s Secret from the key/value pairs you give as input, encrypts it using `kubeseal` and finally applies it to the cluster.
In subsequent runs it will check if the object still exists in the cluster or if the contents have changed and act accordingly.


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

provider "sealedsecrets" {
  kubernetes {
    host                   = data.aws_eks_cluster.cluster.endpoint
    cluster_ca_certificate = base64decode(data.aws_eks_cluster.cluster.certificate_authority.0.data)
    token                  = data.aws_eks_cluster_auth.cluster.token
  }
}

resource "sealedsecrets_secret" "my_secret" {
  type                 = "Opaque"
  name                 = "pg-credentials"
  namespace            = "apps"
  controller_name      = "sealed-secret-controller"
  controller_namespace = "kube-system"
  depends_on           = [kubernetes_namespace.example_ns]
  secrets = {
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
- `secrets` - Key/value pairs to populate the secret

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
