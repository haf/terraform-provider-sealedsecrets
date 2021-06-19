terraform {
  required_providers {
    sealedsecrets = {
      source  = "haf/sealedsecrets"
      version = "0.2.1"
    }
  }
}

provider "sealedsecrets" {
  # optional
  kubectl_bin = "/usr/local/bin/kubectl"

  # optional
  kubeseal_bin = "/usr/local/bin/kubeseal"
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
