package kubeseal

import (
	"bytes"
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"

	ssv1alpha1 "github.com/bitnami-labs/sealed-secrets/pkg/apis/sealed-secrets/v1alpha1"
	"github.com/bitnami-labs/sealed-secrets/pkg/multidocyaml"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeserializer "k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/util/cert"

	"github.com/haf/terraform-provider-sealedsecrets/utils/kubectl"
)

func prettyEncoder(codecs runtimeserializer.CodecFactory, mediaType string, gv runtime.GroupVersioner) (runtime.Encoder, error) {
	info, ok := runtime.SerializerInfoForMediaType(codecs.SupportedMediaTypes(), mediaType)
	if !ok {
		return nil, fmt.Errorf("binary can't serialize %s", mediaType)
	}

	prettyEncoder := info.PrettySerializer
	if prettyEncoder == nil {
		prettyEncoder = info.Serializer
	}

	enc := codecs.EncoderForVersion(prettyEncoder, gv)
	return enc, nil
}

func readSecret(codec runtime.Decoder, r io.Reader) (*v1.Secret, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	if err := multidocyaml.EnsureNotMultiDoc(data); err != nil {
		return nil, err
	}

	var ret v1.Secret
	if err = runtime.DecodeInto(codec, data, &ret); err != nil {
		return nil, err
	}

	return &ret, nil
}

const proxyGetHelp = `kubectl proxy --port 8080 & curl -H "accept:application/json" localhost:8080/api/v1/namespaces/kube-system/services/sealed-secrets-controller/proxy/v1/cert.pem`

/**
resource "google_compute_firewall" "kubeseal" {
  project     = "sample-project"
  name        = "gke-kubeseal-allow-http"
  network     = var.network
  target_tags = local.target_tags
  source_ranges = [
    var.master_ipv4_cidr_block
  ]

  allow {
    protocol = "tcp"
    ports    = ["8080"]
  }
}
*/

func FetchCertificate(controllerName string, controllerNamespace string, kubeProvider *kubectl.KubeProvider) (io.ReadCloser, error) {
	log.Printf("in FetchCertificate, client-go rest.Config: %v\n", &kubeProvider.RestConfig)

	kubeProvider.RestConfig.AcceptContentTypes = "application/x-pem-file, */*"
	restClient, err := corev1.NewForConfig(&kubeProvider.RestConfig)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// https://pkg.go.dev/k8s.io/client-go@v0.21.2/kubernetes/typed/core/v1?utm_source=gopls#ServiceExpansion.ProxyGet
	f, err := restClient.
		Services(controllerNamespace).
		ProxyGet("http", controllerName, "", "/v1/cert.pem", nil).
		Stream(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch certificate, ns=%q, svc=%q — if this is a timeout, make sure this succeeds first: %q — actual error: %v", controllerNamespace, controllerName, proxyGetHelp, err)
	}

	return f, nil
}

func ParseKey(r io.Reader) (*rsa.PublicKey, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	certs, err := cert.ParseCertsPEM(data)
	if err != nil {
		return nil, err
	}

	// ParseCertsPem returns error if len(certs) == 0, but best to be sure...
	if len(certs) == 0 {
		return nil, errors.New("failed to read any certificates")
	}

	cert, ok := certs[0].PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("expected RSA public key but found %v", certs[0].PublicKey)
	}

	return cert, nil
}

func Seal(in io.Reader, pubKey *rsa.PublicKey, scope ssv1alpha1.SealingScope, allowEmptyData bool) (string, error) {
	sealedSecretManifest := new(bytes.Buffer)
	codecs := scheme.Codecs

	secret, err := readSecret(codecs.UniversalDecoder(), in)
	if err != nil {
		return "", err
	}

	if len(secret.Data) == 0 && len(secret.StringData) == 0 && !allowEmptyData {
		return "", fmt.Errorf("secret.data is empty in input Secret, assuming this is an error and aborting. To work with empty data, --allow-empty-data can be used")
	}

	if secret.GetName() == "" {
		return "", fmt.Errorf("missing metadata.name in input Secret")
	}

	if scope != ssv1alpha1.DefaultScope {
		secret.Annotations = ssv1alpha1.UpdateScopeAnnotations(secret.Annotations, scope)
	}

	// Strip read-only server-side ObjectMeta (if present)
	secret.SetSelfLink("")
	secret.SetUID("")
	secret.SetResourceVersion("")
	secret.Generation = 0
	secret.SetCreationTimestamp(metav1.Time{})
	secret.SetDeletionTimestamp(nil)
	secret.DeletionGracePeriodSeconds = nil

	ssecret, err := ssv1alpha1.NewSealedSecret(codecs, pubKey, secret)
	if err != nil {
		return "", err
	}

	var contentType = runtime.ContentTypeYAML

	prettyEnc, err := prettyEncoder(codecs, contentType, ssv1alpha1.SchemeGroupVersion)
	if err != nil {
		return "", err
	}
	buf, err := runtime.Encode(prettyEnc, ssecret)
	if err != nil {
		return "", err
	}

	sealedSecretManifest.Write(buf)
	return sealedSecretManifest.String(), nil
}
