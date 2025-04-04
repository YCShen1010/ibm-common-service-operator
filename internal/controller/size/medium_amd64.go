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

package size

const Medium = `
- name: ibm-cert-manager-operator
  spec:
    certManager:
      certManagerCAInjector:
        resources:
          limits:
            cpu: 100m
            memory: 1000Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 30m
            memory: 500Mi
      certManagerController:
        resources:
          limits:
            cpu: 80m
            memory: 1010Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 20m
            memory: 510Mi
      certManagerWebhook:
        resources:
          limits:
            cpu: 60m
            memory: 100Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 30m
            memory: 40Mi
- name: ibm-mongodb-operator
  spec:
    mongoDB:
      replicas: 3
      resources:
        limits:
          cpu: 2000m
          memory: 2Gi
        requests:
          ephemeral-storage: 256Mi
          cpu: 500m
          memory: 2Gi
      metrics:
        resources:
          requests:
            ephemeral-storage: 256Mi
            cpu: 100m
            memory: 300Mi
          limits:
            cpu: 1000m
            memory: 350Mi
- name: common-service-postgresql
  resources:
  - apiVersion: postgresql.k8s.enterprisedb.io/v1
    kind: Cluster
    name: common-service-db
    data:
      spec:
        instances: 2
        resources:
          limits:
            cpu: 500m
            memory: 1024Mi
            ephemeral-storage: 512Mi
          requests:
            ephemeral-storage: 128Mi
            cpu: 150m
            memory: 384Mi
        postgresql:
          parameters:
            max_connections: "750"
            shared_buffers: 96MB
- name: ibm-im-mongodb-operator-v4.0
  spec:
    mongoDB:
      replicas: 3
      resources:
        limits:
          cpu: 2000m
          memory: 2Gi
        requests:
          ephemeral-storage: 256Mi
          cpu: 500m
          memory: 2Gi
      metrics:
        resources:
          requests:
            ephemeral-storage: 256Mi
            cpu: 100m
            memory: 300Mi
          limits:
            cpu: 1000m
            memory: 350Mi
- name: ibm-im-mongodb-operator-v4.1
  spec:
    mongoDB:
      replicas: 3
      resources:
        limits:
          cpu: 2000m
          memory: 2Gi
        requests:
          ephemeral-storage: 256Mi
          cpu: 500m
          memory: 2Gi
      metrics:
        resources:
          requests:
            ephemeral-storage: 256Mi
            cpu: 100m
            memory: 300Mi
          limits:
            cpu: 1000m
            memory: 350Mi
- name: ibm-im-mongodb-operator-v4.2
  spec:
    mongoDB:
      replicas: 3
      resources:
        limits:
          cpu: 2000m
          memory: 2Gi
        requests:
          ephemeral-storage: 256Mi
          cpu: 500m
          memory: 2Gi
      metrics:
        resources:
          requests:
            ephemeral-storage: 256Mi
            cpu: 100m
            memory: 300Mi
          limits:
            cpu: 1000m
            memory: 350Mi
- name: ibm-im-mongodb-operator
  spec:
    mongoDB:
      replicas: 3
      resources:
        limits:
          cpu: 2000m
          memory: 2Gi
        requests:
          ephemeral-storage: 256Mi
          cpu: 500m
          memory: 2Gi
      metrics:
        resources:
          requests:
            ephemeral-storage: 256Mi
            cpu: 100m
            memory: 300Mi
          limits:
            cpu: 1000m
            memory: 350Mi
- name: ibm-iam-operator
  spec:
    authentication:
      replicas: 2
      auditService:
        resources:
          limits:
            cpu: 20m
            memory: 40Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 10m
            memory: 20Mi
      authService:
        resources:
          limits:
            cpu: 1000m
            memory: 1090Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 600m
            memory: 540Mi
      clientRegistration:
        resources:
          limits:
            cpu: 1000m
            memory: 50Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 20m
            memory: 50Mi
      identityManager:
        resources:
          limits:
            cpu: 1000m
            memory: 1260Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 260m
            memory: 240Mi
      identityProvider:
        resources:
          limits:
            cpu: 1000m
            memory: 920Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 570m
            memory: 250Mi
    oidcclientwatcher:
      replicas: 1
      resources:
        limits:
          cpu: 1000m
          memory: 325Mi
        requests:
          ephemeral-storage: 256Mi
          cpu: 20m
          memory: 40Mi
    pap:
      auditService:
        resources:
          limits:
            cpu: 20m
            memory: 40Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 10m
            memory: 20Mi
      papService:
        resources:
          limits:
            cpu: 1000m
            memory: 600Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 30m
            memory: 190Mi
      replicas: 2
    policycontroller:
      replicas: 1
      resources:
        limits:
          cpu: 1000m
          memory: 80Mi
        requests:
          ephemeral-storage: 256Mi
          cpu: 20m
          memory: 40Mi
    policydecision:
      auditService:
        resources:
          limits:
            cpu: 20m
            memory: 40Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 10m
            memory: 20Mi
      resources:
        limits:
          cpu: 1000m
          memory: 160Mi
        requests:
          ephemeral-storage: 256Mi
          cpu: 70m
          memory: 30Mi
      replicas: 2
    secretwatcher:
      resources:
        limits:
          cpu: 1000m
          memory: 220Mi
        requests:
          ephemeral-storage: 256Mi
          cpu: 20m
          memory: 50Mi
      replicas: 1
    securityonboarding:
      replicas: 1
      resources:
        limits:
          cpu: 1000m
          memory: 50Mi
        requests:
          ephemeral-storage: 256Mi
          cpu: 20m
          memory: 50Mi
      iamOnboarding:
        resources:
          limits:
            cpu: 1000m
            memory: 1024Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 20m
            memory: 64Mi
- name: ibm-im-operator
  spec:
    authentication:
      replicas: 2
      authService:
        resources:
          limits:
            cpu: 1000m
            memory: 1090Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 600m
            memory: 540Mi
      clientRegistration:
        resources:
          limits:
            cpu: 1000m
            memory: 50Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 20m
            memory: 50Mi
      identityManager:
        resources:
          limits:
            cpu: 1000m
            memory: 1260Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 260m
            memory: 240Mi
      identityProvider:
        resources:
          limits:
            cpu: 1000m
            memory: 920Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 570m
            memory: 250Mi
- name: ibm-im-operator-v4.0
  spec:
    authentication:
      replicas: 2
      authService:
        resources:
          limits:
            cpu: 1000m
            memory: 1090Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 600m
            memory: 540Mi
      clientRegistration:
        resources:
          limits:
            cpu: 1000m
            memory: 50Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 20m
            memory: 50Mi
      identityManager:
        resources:
          limits:
            cpu: 1000m
            memory: 1260Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 260m
            memory: 240Mi
      identityProvider:
        resources:
          limits:
            cpu: 1000m
            memory: 920Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 570m
            memory: 250Mi
- name: ibm-im-operator-v4.1
  spec:
    authentication:
      replicas: 2
      authService:
        resources:
          limits:
            cpu: 1000m
            memory: 1090Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 600m
            memory: 540Mi
      clientRegistration:
        resources:
          limits:
            cpu: 1000m
            memory: 50Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 20m
            memory: 50Mi
      identityManager:
        resources:
          limits:
            cpu: 1000m
            memory: 1260Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 260m
            memory: 240Mi
      identityProvider:
        resources:
          limits:
            cpu: 1000m
            memory: 920Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 570m
            memory: 250Mi
- name: ibm-im-operator-v4.2
  spec:
    authentication:
      replicas: 2
      authService:
        resources:
          limits:
            cpu: 1000m
            memory: 1090Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 600m
            memory: 540Mi
      clientRegistration:
        resources:
          limits:
            cpu: 1000m
            memory: 50Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 20m
            memory: 50Mi
      identityManager:
        resources:
          limits:
            cpu: 1000m
            memory: 1260Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 260m
            memory: 240Mi
      identityProvider:
        resources:
          limits:
            cpu: 1000m
            memory: 920Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 570m
            memory: 250Mi
- name: ibm-im-operator-v4.3
  spec:
    authentication:
      replicas: 2
      authService:
        resources:
          limits:
            cpu: 1000m
            memory: 1090Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 600m
            memory: 540Mi
      clientRegistration:
        resources:
          limits:
            cpu: 1000m
            memory: 50Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 20m
            memory: 50Mi
      identityManager:
        resources:
          limits:
            cpu: 1000m
            memory: 1260Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 260m
            memory: 240Mi
      identityProvider:
        resources:
          limits:
            cpu: 1000m
            memory: 920Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 570m
            memory: 250Mi
- name: ibm-im-operator-v4.4
  spec:
    authentication:
      replicas: 2
      authService:
        resources:
          limits:
            cpu: 1000m
            memory: 1090Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 600m
            memory: 540Mi
      clientRegistration:
        resources:
          limits:
            cpu: 1000m
            memory: 50Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 20m
            memory: 50Mi
      identityManager:
        resources:
          limits:
            cpu: 1000m
            memory: 1260Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 260m
            memory: 240Mi
      identityProvider:
        resources:
          limits:
            cpu: 1000m
            memory: 920Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 570m
            memory: 250Mi
- name: ibm-im-operator-v4.5
  spec:
    authentication:
      replicas: 2
      authService:
        resources:
          limits:
            cpu: 1000m
            memory: 1090Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 600m
            memory: 540Mi
      clientRegistration:
        resources:
          limits:
            cpu: 1000m
            memory: 50Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 20m
            memory: 50Mi
      identityManager:
        resources:
          limits:
            cpu: 1000m
            memory: 1260Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 260m
            memory: 240Mi
      identityProvider:
        resources:
          limits:
            cpu: 1000m
            memory: 920Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 570m
            memory: 250Mi
- name: ibm-im-operator-v4.6
  spec:
    authentication:
      replicas: 2
      authService:
        resources:
          limits:
            cpu: 1000m
            memory: 1090Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 600m
            memory: 540Mi
      clientRegistration:
        resources:
          limits:
            cpu: 1000m
            memory: 50Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 20m
            memory: 50Mi
      identityManager:
        resources:
          limits:
            cpu: 1000m
            memory: 1260Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 260m
            memory: 240Mi
      identityProvider:
        resources:
          limits:
            cpu: 1000m
            memory: 920Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 570m
            memory: 250Mi
- name: ibm-management-ingress-operator
  spec:
    managementIngress:
      replicas: 2
      resources:
        requests:
          ephemeral-storage: 256Mi
          cpu: 70m
          memory: 70Mi
        limits:
          cpu: 1000m
          memory: 760Mi
- name: ibm-ingress-nginx-operator
  spec:
    nginxIngress:
      ingress:
        replicas: 2
        resources:
          requests:
            ephemeral-storage: 256Mi
            cpu: 100m
            memory: 120Mi
          limits:
            cpu: 1000m
            memory: 860Mi
      defaultBackend:
        replicas: 1
        resources:
          requests:
            ephemeral-storage: 256Mi
            cpu: 20m
            memory: 20Mi
          limits:
            cpu: 50m
            memory: 100Mi
      kubectl:
        resources:
          requests:
            ephemeral-storage: 256Mi
            memory: 150Mi
            cpu: 30m
          limits:
            memory: 150Mi
            cpu: 50m
- name: ibm-licensing-operator
  spec:
    IBMLicensing:
      resources:
        requests:
          ephemeral-storage: 256Mi
          cpu: 200m
          memory: 430Mi
        limits:
          cpu: 300m
          memory: 850Mi
    IBMLicenseServiceReporter:
      databaseContainer:
        resources:
          requests:
            ephemeral-storage: 256Mi
            cpu: 200m
            memory: 256Mi
          limits:
            cpu: 300m
            memory: 300Mi
      receiverContainer:
        resources:
          requests:
            ephemeral-storage: 256Mi
            cpu: 200m
            memory: 256Mi
          limits:
            cpu: 300m
            memory: 384Mi
- name: ibm-commonui-operator
  spec:
    commonWebUI:
      replicas: 2
      resources:
        requests:
          ephemeral-storage: 256Mi
          memory: 480Mi
          cpu: 450m
        limits:
          memory: 660Mi
          cpu: 1000m
- name: ibm-idp-config-ui-operator
  spec:
    commonWebUI:
      replicas: 2
      resources:
        requests:
          ephemeral-storage: 256Mi
          memory: 480Mi
          cpu: 450m
        limits:
          memory: 660Mi
          cpu: 1000m
- name: ibm-idp-config-ui-operator-v4.0
  spec:
    commonWebUI:
      replicas: 2
      resources:
        requests:
          ephemeral-storage: 256Mi
          memory: 480Mi
          cpu: 450m
        limits:
          memory: 660Mi
          cpu: 1000m
- name: ibm-idp-config-ui-operator-v4.1
  spec:
    commonWebUI:
      replicas: 2
      resources:
        requests:
          ephemeral-storage: 256Mi
          memory: 480Mi
          cpu: 450m
        limits:
          memory: 660Mi
          cpu: 1000m
- name: ibm-idp-config-ui-operator-v4.2
  spec:
    commonWebUI:
      replicas: 2
      resources:
        requests:
          ephemeral-storage: 256Mi
          memory: 480Mi
          cpu: 450m
        limits:
          memory: 660Mi
          cpu: 1000m
- name: ibm-idp-config-ui-operator-v4.3
  spec:
    commonWebUI:
      replicas: 2
      resources:
        requests:
          ephemeral-storage: 256Mi
          memory: 480Mi
          cpu: 450m
        limits:
          memory: 660Mi
          cpu: 1000m
- name: ibm-idp-config-ui-operator-v4.4
  spec:
    commonWebUI:
      replicas: 2
      resources:
        requests:
          ephemeral-storage: 256Mi
          memory: 480Mi
          cpu: 450m
        limits:
          memory: 660Mi
          cpu: 1000m
- name: ibm-idp-config-ui-operator-v4.5
  spec:
    commonWebUI:
      replicas: 2
      resources:
        requests:
          ephemeral-storage: 256Mi
          memory: 480Mi
          cpu: 450m
        limits:
          memory: 660Mi
          cpu: 1000m
- name: ibm-platform-api-operator
  spec:
    platformApi:
      auditService:
        resources:
          limits:
            cpu: 25m
            memory: 50Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 20m
            memory: 30Mi
      platformApi:
        resources:
          limits:
            cpu: 25m
            memory: 60Mi
          requests:
            ephemeral-storage: 256Mi
            cpu: 20m
            memory: 30Mi
      replicas: 1
- name: ibm-healthcheck-operator
  spec:
    healthService:
      memcached:
        replicas: 1
        resources:
          requests:
            ephemeral-storage: 256Mi
            memory: 50Mi
            cpu: 20m
          limits:
            memory: 100Mi
            cpu: 200m
      healthService:
        replicas: 1
        resources:
          requests:
            ephemeral-storage: 256Mi
            memory: 125Mi
            cpu: 20m
          limits:
            memory: 250Mi
            cpu: 200m
- name: ibm-auditlogging-operator
  spec:
    auditLogging:
      fluentd:
        resources:
          requests:
            ephemeral-storage: 256Mi
            cpu: 35m
            memory: 20Mi
          limits:
            cpu: 65m
            memory: 200Mi
- name: ibm-monitoring-grafana-operator
  spec:
    grafana:
      grafanaConfig:
        resources:
          requests:
            ephemeral-storage: 256Mi
            cpu: 20m
            memory: 60Mi
          limits:
            cpu: 150m
            memory: 230Mi
      dashboardConfig:
        resources:
          requests:
            ephemeral-storage: 256Mi
            cpu: 25m
            memory: 140Mi
          limits:
            cpu: 70m
            memory: 170Mi
      routerConfig:
        resources:
          requests:
            ephemeral-storage: 256Mi
            cpu: 20m
            memory: 20Mi
          limits:
            cpu: 70m
            memory: 80Mi
- name: ibm-apicatalog
  spec:
    apicatalogmanager:
      profile: medium
- name: edb-keycloak
  resources:
  - apiVersion: postgresql.k8s.enterprisedb.io/v1
    kind: Cluster
    name: keycloak-edb-cluster
    data:
      spec:
        instances: 2
        resources:
          limits:
            cpu: 500m
            memory: 1024Mi
          requests:
            cpu: 500m
            memory: 1024Mi
- name: keycloak-operator
  resources:
  - apiVersion: k8s.keycloak.org/v2alpha1
    kind: Keycloak
    name: cs-keycloak
    data:
      spec:
        instances: 2
        resources:
          limits:
            cpu: 1000m
            memory: 1Gi
            ephemeral-storage: 512Mi
          requests:
            cpu: 1000m
            memory: 1Gi
            ephemeral-storage: 256Mi
        unsupported:
          podTemplate:
            spec:
              containers:
                - resources:
                    limits:
                      cpu: 1000m
                      memory: 1Gi
                      ephemeral-storage: 512Mi
                    requests:
                      cpu: 1000m
                      memory: 1Gi
                      ephemeral-storage: 256Mi
`
