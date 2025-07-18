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

package main

import (
	"flag"
	"os"
	"strings"

	olmv1 "github.com/operator-framework/api/pkg/operators/v1"
	olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	operatorsv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"
	admv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	nssv1 "github.com/IBM/ibm-namespace-scope-operator/v4/api/v1"
	ssv1 "github.com/IBM/ibm-secretshare-operator/api/v1"
	odlm "github.com/IBM/operand-deployment-lifecycle-manager/v4/api/v1alpha1"

	certmanagerv1 "github.com/ibm/ibm-cert-manager-operator/apis/cert-manager/v1"

	operatorv3 "github.com/IBM/ibm-common-service-operator/v4/api/v3"
	controllers "github.com/IBM/ibm-common-service-operator/v4/internal/controller"
	"github.com/IBM/ibm-common-service-operator/v4/internal/controller/bootstrap"
	certmanagerv1controllers "github.com/IBM/ibm-common-service-operator/v4/internal/controller/cert-manager"
	util "github.com/IBM/ibm-common-service-operator/v4/internal/controller/common"
	"github.com/IBM/ibm-common-service-operator/v4/internal/controller/constant"
	"github.com/IBM/ibm-common-service-operator/v4/internal/controller/goroutines"
	commonservicewebhook "github.com/IBM/ibm-common-service-operator/v4/internal/controller/webhooks/commonservice"
	operandrequestwebhook "github.com/IBM/ibm-common-service-operator/v4/internal/controller/webhooks/operandrequest"
	// +kubebuilder:scaffold:imports
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(odlm.AddToScheme(scheme))
	utilruntime.Must(nssv1.AddToScheme(scheme))
	utilruntime.Must(ssv1.AddToScheme(scheme))
	utilruntime.Must(operatorv3.AddToScheme(scheme))
	utilruntime.Must(admv1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme

	utilruntime.Must(olmv1alpha1.AddToScheme(scheme))
	utilruntime.Must(olmv1.AddToScheme(scheme))
	utilruntime.Must(operatorsv1.AddToScheme(scheme))
	utilruntime.Must(certmanagerv1.AddToScheme(scheme))
}

func main() {
	klog.InitFlags(nil)
	defer klog.Flush()
	var metricsAddr string
	var probeAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	options := ctrl.Options{
		Scheme:                 scheme,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "ab89bbb1.ibm.com",
	}

	watchNamespace := util.GetWatchNamespace()

	// var NewCache cache.NewCacheFunc
	watchNamespaceList := strings.Split(watchNamespace, ",")
	options = util.NewCSCache(watchNamespaceList, options)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), options)
	if err != nil {
		klog.Errorf("Unable to start manager: %v", err)
		os.Exit(1)
	}

	operatorNs, err := util.GetOperatorNamespace()
	klog.Infof("Identifying Common Service Operator Role in the namespace %s", operatorNs)
	if err != nil {
		klog.Errorf("Failed to get operatorNs: %v", err)
		os.Exit(1)
	}

	cpfsNs, err := bootstrap.IdentifyCPFSNs(mgr.GetAPIReader(), operatorNs)
	if err != nil {
		klog.Errorf("Failed to get Common Service deployed namespace: %v", err)
		os.Exit(1)
	}
	// If Common Service Operator Namespace is not in the same as .spec.operatorNamespace(cpfsNs) in default CS CR,
	// this Common Service Operator is not in the operatorNamespace(cpfsNs) under this tenant, and goes dormant.
	if operatorNs == cpfsNs {
		// New bootstrap Object
		var bs *bootstrap.Bootstrap
		if os.Getenv("NO_OLM") == "true" {
			bs, err = bootstrap.NewNonOLMBootstrap(mgr)
			if err != nil {
				klog.Errorf("No olm Bootstrap failed: %v", err)
				os.Exit(1)
			}
		} else {
			bs, err = bootstrap.NewBootstrap(mgr)
			if err != nil {
				klog.Errorf("Bootstrap failed: %v", err)
				os.Exit(1)
			}
		}

		if err := bs.CleanupWebhookResources(); err != nil {
			klog.Errorf("Cleanup Webhook Resources failed: %v", err)
			os.Exit(1)
		}
		klog.Infof("Setup commonservice manager")
		if err = (&controllers.CommonServiceReconciler{
			Bootstrap: bs,
			Scheme:    mgr.GetScheme(),
			Recorder:  mgr.GetEventRecorderFor("commonservice-controller"),
		}).SetupWithManager(mgr); err != nil {
			klog.Errorf("Unable to create controller CommonService: %v", err)
			os.Exit(1)
		}

		// Create CS CR
		klog.Infof("Start go routines")
		if os.Getenv("NO_OLM") == "true" {
			go goroutines.WaitToCreateCsCRNoOLM(bs)
		} else {
			go goroutines.WaitToCreateCsCR(bs)
		}
		// Delete Keycloak Cert
		go goroutines.CleanupResources(bs)

		// check if cert-manager CRD does not exist, then skip cert-manager related controllers initialization
		exist, err := bs.CheckCRD(constant.CertManagerAPIGroupVersionV1, "Certificate")
		if err != nil {
			klog.Errorf("Failed to check if cert-manager CRD exists: %v", err)
			os.Exit(1)
		}
		if !exist && err == nil {
			klog.Infof("cert-manager CRD does not exist, skip cert-manager related controllers initialization")
		} else if exist && err == nil {
			if err = (&certmanagerv1controllers.CertificateRefreshReconciler{
				Client: mgr.GetClient(),
				Scheme: mgr.GetScheme(),
			}).SetupWithManager(mgr); err != nil {
				klog.Error(err, "unable to create controller", "controller", "CertificateRefresh")
				os.Exit(1)
			}
			if err = (&certmanagerv1controllers.PodRefreshReconciler{
				Client: mgr.GetClient(),
				Scheme: mgr.GetScheme(),
			}).SetupWithManager(mgr); err != nil {
				klog.Error(err, "unable to create controller", "controller", "PodRefresh")
				os.Exit(1)
			}
			if err = (&certmanagerv1controllers.V1AddLabelReconciler{
				Client: mgr.GetClient(),
				Scheme: mgr.GetScheme(),
			}).SetupWithManager(mgr); err != nil {
				klog.Error(err, "unable to create controller", "controller", "V1AddLabel")
				os.Exit(1)
			}
		}
	} else {
		klog.Infof("Common Service Operator goes dormant in the namespace %s", operatorNs)
		klog.Infof("Common Service Operator in the namespace %s takes charge of resource management", cpfsNs)
	}

	// Start up the webhook server
	if os.Getenv("ENABLE_WEBHOOKS") != "false" {
		if err = (&commonservicewebhook.Defaulter{
			Client:    mgr.GetClient(),
			Reader:    mgr.GetAPIReader(),
			IsDormant: operatorNs != cpfsNs,
		}).SetupWebhookWithManager(mgr); err != nil {
			klog.Errorf("Unable to create CommonService webhook: %v", err)
			os.Exit(1)
		}

		if err = (&operandrequestwebhook.Defaulter{
			Client:    mgr.GetClient(),
			Reader:    mgr.GetAPIReader(),
			IsDormant: operatorNs != cpfsNs,
		}).SetupWebhookWithManager(mgr); err != nil {
			klog.Errorf("Unable to create OperandRequest webhook: %v", err)
			os.Exit(1)
		}
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		klog.Errorf("unable to set up health check: %v", err)
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		klog.Errorf("unable to set up ready check: %v", err)
		os.Exit(1)
	}

	klog.Info("Starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		klog.Errorf("Problem running manager: %v", err)
		os.Exit(1)
	}

}
