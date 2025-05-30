//
// Copyright 2022 IBM Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package constant

// SecretWatchLabel is a string of secrets that watched by cert manager operator labels
const SecretWatchLabel string = "operator.ibm.com/watched-by-cert-manager"

// Labels and Annotations added by this operator
const (
	OperatorGeneratedAnno   = "ibm-cert-manager-operator-generated"
	ProperV1Label           = "ibm-cert-manager-operator/conditionally-generated-v1"
	RefreshCALabel          = "ibm-cert-manager-operator/refresh-ca-chain"
	ManageCertRotationLabel = "manage-cert-rotation"
)

var (
	CertManagerAPIGroupVersionV1Alpha1 = "certmanager.k8s.io/v1alpha1"
	CertManagerAPIGroupVersionV1       = "cert-manager.io/v1"
	CertManagerKinds                   = []string{"Issuer", "Certificate"}
	CertManagerIssuers                 = []string{CSSSIssuer, CSCAIssuer}
	CertManagerCerts                   = []string{CSCACert}
	KeycloakCert                       = "cs-keycloak-tls-cert"
)

// CSCAIssuer is the CR of cs-ca-issuer
const CSCAIssuer = `
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  labels:
    app.kubernetes.io/instance: cs-ca-issuer
    app.kubernetes.io/managed-by: cert-manager-controller
    app.kubernetes.io/name: Issuer
    operator.ibm.com/managedByCsOperator: "true"
  name: cs-ca-issuer
  namespace: "placeholder"
spec:
  ca:
    secretName: cs-ca-certificate-secret
`

// CSSSIsuuer is the CR of cs-ss-issuer
const CSSSIssuer = `
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  labels:
    app.kubernetes.io/instance: cs-ss-issuer
    app.kubernetes.io/managed-by: cert-manager-controller
    app.kubernetes.io/name: Issuer
    operator.ibm.com/managedByCsOperator: "true"
  name: cs-ss-issuer
  namespace: "placeholder"
spec:
  selfSigned: {}
`

// CSCACert is the CR of cs-ca-certificate
const CSCACert = `
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  labels:
    app.kubernetes.io/instance: cs-ca-certificate
    app.kubernetes.io/managed-by: cert-manager-controller
    app.kubernetes.io/name: Certificate
    operator.ibm.com/managedByCsOperator: 'true'
    ibm-cert-manager-operator/refresh-ca-chain: 'true'
    manage-cert-rotation: 'true'
  name: cs-ca-certificate
  namespace: "placeholder"
spec:
  secretName: cs-ca-certificate-secret
  secretTemplate:
    labels:
      ibm-cert-manager-operator/refresh-ca-chain: 'true'
  issuerRef:
    name: cs-ss-issuer
    kind: Issuer
  commonName: cs-ca-certificate
  isCA: true
  duration: 17520h0m0s
  renewBefore: 5840h0m0s
`

const KeycloakCertTemplate = `
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: cs-keycloak-tls-cert
  namespace: {{ .ServicesNs }}
spec:
  commonName: cs-keycloak-service
  dnsNames:
      - cs-keycloak-service
      - cs-keycloak-service.{{ .ServicesNs }}
      - cs-keycloak-service.{{ .ServicesNs }}.svc
      - cs-keycloak-service.{{ .ServicesNs }}.svc.cluster.local
  issuerRef:
      kind: Issuer
      name: cs-ca-issuer
  secretName: cs-keycloak-tls-secret
`
