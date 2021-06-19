terraform {
  required_providers {
    sealedsecrets = {
      source  = "haf/sealedsecrets"
      version = "0.2.4"
    }
  }
}

provider "sealedsecrets" {
  # optional
  kubectl_bin = "/usr/local/bin/kubectl"

  # optional
  kubeseal_bin = "/usr/local/bin/kubeseal"

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
