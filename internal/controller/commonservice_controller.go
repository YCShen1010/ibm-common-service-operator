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

package controllers

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"

	olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	apiv3 "github.com/IBM/ibm-common-service-operator/v4/api/v3"
	"github.com/IBM/ibm-common-service-operator/v4/internal/controller/bootstrap"
	util "github.com/IBM/ibm-common-service-operator/v4/internal/controller/common"
	"github.com/IBM/ibm-common-service-operator/v4/internal/controller/configurationcollector"
	"github.com/IBM/ibm-common-service-operator/v4/internal/controller/constant"
	odlm "github.com/IBM/operand-deployment-lifecycle-manager/v4/api/v1alpha1"
)

// CommonServiceReconciler reconciles a CommonService object
type CommonServiceReconciler struct {
	*bootstrap.Bootstrap
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

func (r *CommonServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	klog.Infof("Reconciling CommonService: %s", req.NamespacedName)

	// Fetch the CommonService instance
	instance := &apiv3.CommonService{}
	if req.Name == constant.MasterCR && util.Contains(strings.Split(r.Bootstrap.CSData.WatchNamespaces, ","), req.Namespace) && req.Namespace != r.Bootstrap.CSData.OperatorNs {
		if err := r.Bootstrap.Client.Get(ctx, req.NamespacedName, instance); err != nil {
			if errors.IsNotFound(err) {
				klog.Infof("Finished reconciling to delete CommonService: %s/%s", req.NamespacedName.Namespace, req.NamespacedName.Name)
			}
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		return r.ReconcileNonConfigurableCR(ctx, instance)
	}

	if err := r.Reader.Get(ctx, req.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			if err := r.handleDelete(ctx); err != nil {
				return ctrl.Result{}, err
			}
			// Generate Issuer and Certificate CR
			if err := r.Bootstrap.DeployCertManagerCR(); err != nil {
				return ctrl.Result{}, err
			}
			klog.Infof("Finished reconciling to delete CommonService: %s/%s", req.NamespacedName.Namespace, req.NamespacedName.Name)
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !instance.Spec.License.Accept {
		klog.Error("Accept license by changing .spec.license.accept to true in the CommonService CR. Operator will not proceed until then")
	}

	if os.Getenv("NO_OLM") == "true" {
		klog.Infof("Reconciling CommonService: %s in No OLM environment", req.NamespacedName)
		return r.NoOLMReconcile(ctx, req, instance)
	}

	// If the CommonService CR is not paused, continue to reconcile
	if !r.reconcilePauseRequest(instance) {
		if r.checkNamespace(req.NamespacedName.String()) {
			return r.ReconcileMasterCR(ctx, instance)
		}
		return r.ReconcileGeneralCR(ctx, instance)
	}
	// If the CommonService CR is paused, update the status to pending
	if err := r.updatePhase(ctx, instance, apiv3.CRPending); err != nil {
		klog.Errorf("Fail to reconcile %s/%s: %v", instance.Namespace, instance.Name, err)
		return ctrl.Result{}, err
	}
	klog.Infof("%s/%s is in pending status due to pause request", instance.Namespace, instance.Name)
	return ctrl.Result{}, nil
}

func (r *CommonServiceReconciler) ReconcileMasterCR(ctx context.Context, instance *apiv3.CommonService) (ctrl.Result, error) {

	var statusErr error
	// Defer to Set error/ready/warning condition
	defer func() {
		if err := r.Bootstrap.CheckWarningCondition(instance); err != nil {
			klog.Warning(err)
			return
		}
		if statusErr != nil {
			instance.SetErrorCondition(constant.MasterCR, apiv3.ConditionTypeError, corev1.ConditionTrue, apiv3.ConditionReasonError, statusErr.Error())
		} else {
			instance.SetReadyCondition(constant.KindCR, apiv3.ConditionTypeReady, corev1.ConditionTrue)
		}
		if err := r.Client.Status().Update(ctx, instance); err != nil {
			klog.Warning(err)
			return
		}
	}()

	originalInstance := instance.DeepCopy()

	operatorDeployed, servicesDeployed := r.Bootstrap.CheckDeployStatus(ctx)
	instance.UpdateConfigStatus(&r.Bootstrap.CSData, operatorDeployed, servicesDeployed)

	r.Bootstrap.CSData.CPFSNs = string(instance.Status.ConfigStatus.OperatorNamespace)
	r.Bootstrap.CSData.ServicesNs = string(instance.Status.ConfigStatus.ServicesNamespace)
	r.Bootstrap.CSData.CatalogSourceName = string(instance.Status.ConfigStatus.CatalogName)
	r.Bootstrap.CSData.CatalogSourceNs = string(instance.Status.ConfigStatus.CatalogNamespace)

	var forceUpdateODLMCRs bool
	if !reflect.DeepEqual(originalInstance.Status, instance.Status) {
		forceUpdateODLMCRs = true
	}

	if statusErr = r.Client.Status().Patch(ctx, instance, client.MergeFrom(originalInstance)); statusErr != nil {
		return ctrl.Result{}, fmt.Errorf("error while patching CommonService.Status: %v", statusErr)
	}

	if instance.Status.Phase == "" {
		// Set "Reconciling" condition and "Initializing" for phase
		instance.SetPendingCondition(constant.MasterCR, apiv3.ConditionTypeReconciling, corev1.ConditionTrue, apiv3.ConditionReasonReconcile, apiv3.ConditionMessageReconcile)
		instance.Status.Phase = apiv3.CRInitializing
		if statusErr = r.Client.Status().Update(ctx, instance); statusErr != nil {
			klog.Errorf("Fail to update %s/%s: %v", instance.Namespace, instance.Name, statusErr)
			return ctrl.Result{}, statusErr
		}
	} else {
		if statusErr = r.updatePhase(ctx, instance, apiv3.CRUpdating); statusErr != nil {
			klog.Errorf("Fail to reconcile %s/%s: %v", instance.Namespace, instance.Name, statusErr)
			return ctrl.Result{}, statusErr
		}
	}

	// Creating/updating common-service-maps, skip when installing in AllNamespace Mode
	if r.Bootstrap.CSData.WatchNamespaces != "" {
		cm, err := util.GetCmOfMapCs(r.Reader)
		if err != nil {
			// Create new common-service-maps
			if errors.IsNotFound(err) {
				klog.Infof("Creating common-service-maps ConfigMap in kube-public")
				if err = r.Bootstrap.CreateCsMaps(); err != nil {
					klog.Errorf("Failed to create common-service-maps ConfigMap: %v", err)
					os.Exit(1)
				}
			} else if !errors.IsNotFound(err) {
				klog.Errorf("Failed to get common-service-maps: %v", err)
				os.Exit(1)
			}
		} else {
			// Update common-service-maps
			klog.Infof("Updating common-service-maps ConfigMap in kube-public")
			if err := util.UpdateCsMaps(cm, r.Bootstrap.CSData.WatchNamespaces, r.Bootstrap.CSData.ServicesNs, r.Bootstrap.CSData.OperatorNs); err != nil {
				klog.Errorf("Failed to update common-service-maps: %v", err)
				os.Exit(1)
			}
			// Validate common-service-maps
			if err := util.ValidateCsMaps(cm); err != nil {
				klog.Errorf("Unsupported common-service-maps: %v", err)
				os.Exit(1)
			}
			if err := r.Client.Update(context.TODO(), cm); err != nil {
				klog.Errorf("Failed to update namespaceMapping in common-service-maps: %v", err)
				os.Exit(1)
			}
		}
	} else {
		// check if the servicesNamespace is created
		ns := &corev1.Namespace{}
		if err := r.Reader.Get(ctx, types.NamespacedName{Name: r.Bootstrap.CSData.ServicesNs}, ns); err != nil {
			if errors.IsNotFound(err) {
				klog.Errorf("Not found servicesNamespace %s specified in the common-service CR.", r.Bootstrap.CSData.ServicesNs)
				if err := r.updatePhase(ctx, instance, apiv3.CRFailed); err != nil {
					klog.Error(err)
				}
				klog.Errorf("Fail to reconcile %s/%s: %v", instance.Namespace, instance.Name, err)
				return ctrl.Result{}, err
			}
		}
	}

	typeCorrect, err := r.Bootstrap.CheckClusterType(util.GetServicesNamespace(r.Reader))
	if err != nil {
		klog.Errorf("Failed to verify cluster type  %v", err)
		if err := r.updatePhase(ctx, instance, apiv3.CRFailed); err != nil {
			klog.Error(err)
		}
		klog.Errorf("Fail to reconcile %s/%s: %v", instance.Namespace, instance.Name, err)
		return ctrl.Result{}, err
	}

	if !typeCorrect {
		klog.Error("Cluster type specificed in the ibm-cpp-config isn't correct")
		if statusErr = r.updatePhase(ctx, instance, apiv3.CRFailed); statusErr != nil {
			klog.Error(statusErr)
		}
		klog.Errorf("Fail to reconcile %s/%s: %v", instance.Namespace, instance.Name, statusErr)
		return ctrl.Result{}, statusErr
	}

	// Init common service bootstrap resource
	// Including namespace-scope configmap
	// Deploy OperandConfig and OperandRegistry
	if statusErr = r.Bootstrap.InitResources(instance, forceUpdateODLMCRs); statusErr != nil {
		if statusErr := r.updatePhase(ctx, instance, apiv3.CRFailed); statusErr != nil {
			klog.Error(statusErr)
		}
		klog.Errorf("Fail to reconcile %s/%s: %v", instance.Namespace, instance.Name, statusErr)
		return ctrl.Result{}, statusErr
	}

	// Generate Issuer and Certificate CR
	if statusErr = r.Bootstrap.DeployCertManagerCR(); statusErr != nil {
		klog.Errorf("Failed to deploy cert manager CRs: %v", statusErr)
		if statusErr = r.updatePhase(ctx, instance, apiv3.CRFailed); statusErr != nil {
			klog.Error(statusErr)
		}
		klog.Errorf("Fail to reconcile %s/%s: %v", instance.Namespace, instance.Name, statusErr)
		return ctrl.Result{}, statusErr
	}

	// Apply new configs to CommonService CR
	cs := util.NewUnstructured("operator.ibm.com", "CommonService", "v3")
	if statusErr = r.Bootstrap.Client.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, cs); statusErr != nil {
		klog.Errorf("Fail to reconcile %s/%s: %v", instance.Namespace, instance.Name, statusErr)
		return ctrl.Result{}, statusErr
	}
	// Set "Pending" condition and "Updating" for phase when config CS CR
	instance.SetPendingCondition(constant.MasterCR, apiv3.ConditionTypeReconciling, corev1.ConditionTrue, apiv3.ConditionReasonConfig, apiv3.ConditionMessageConfig)
	instance.Status.Phase = apiv3.CRUpdating
	newConfigs, serviceControllerMapping, statusErr := r.getNewConfigs(cs)
	if statusErr != nil {
		klog.Errorf("Fail to reconcile %s/%s: %v", instance.Namespace, instance.Name, statusErr)
		instance.SetErrorCondition(constant.MasterCR, apiv3.ConditionTypeError, corev1.ConditionTrue, apiv3.ConditionReasonError, statusErr.Error())
		instance.Status.Phase = apiv3.CRFailed
	}

	if statusErr = r.Client.Status().Update(ctx, instance); statusErr != nil {
		klog.Errorf("Fail to update %s/%s: %v", instance.Namespace, instance.Name, statusErr)
		return ctrl.Result{}, statusErr
	}

	var isEqual bool
	if isEqual, statusErr = r.updateOperandConfig(ctx, newConfigs, serviceControllerMapping); statusErr != nil {
		if statusErr := r.updatePhase(ctx, instance, apiv3.CRFailed); statusErr != nil {
			klog.Error(statusErr)
		}
		klog.Errorf("Fail to reconcile %s/%s: %v", instance.Namespace, instance.Name, statusErr)
		return ctrl.Result{}, statusErr
	} else if isEqual {
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Noeffect", fmt.Sprintf("No update, resource sizings in the OperandConfig %s/%s are larger than the profile from CommonService CR %s/%s", r.Bootstrap.CSData.OperatorNs, "common-service", instance.Namespace, instance.Name))
	}

	if statusErr = r.Bootstrap.UpdateEDBUserManaged(); statusErr != nil {
		if statusErr := r.updatePhase(ctx, instance, apiv3.CRFailed); statusErr != nil {
			klog.Error(statusErr)
		}
		klog.Errorf("Fail to reconcile %s/%s: %v", instance.Namespace, instance.Name, statusErr)
		return ctrl.Result{}, statusErr
	}

	if isEqual, statusErr = r.updateOperatorConfig(ctx, instance.Spec.OperatorConfigs); statusErr != nil {
		if statusErr := r.updatePhase(ctx, instance, apiv3.CRFailed); statusErr != nil {
			klog.Error(statusErr)
		}
		klog.Errorf("Fail to reconcile %s/%s: %v", instance.Namespace, instance.Name, statusErr)
		return ctrl.Result{}, statusErr
	} else if isEqual {
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Noeffect", fmt.Sprintf("No update, replica sizings in the OperatorConfig %s/%s are larger than the profile from CommonService CR %s/%s", r.Bootstrap.CSData.OperatorNs, "common-service", instance.Namespace, instance.Name))
	}

	if statusErr = configurationcollector.CreateUpdateConfig(r.Bootstrap); statusErr != nil {
		if statusErr := r.updatePhase(ctx, instance, apiv3.CRFailed); statusErr != nil {
			klog.Error(statusErr)
		}
		klog.Errorf("Fail to reconcile %s/%s: %v", instance.Namespace, instance.Name, statusErr)
		return ctrl.Result{}, statusErr
	}

	// Wait for Postgres Cluster image to be updated
	if statusErr = r.Bootstrap.UpdatePostgresClusterImage(ctx, instance); statusErr != nil {
		klog.Errorf("Failed to update Postgres Cluster image: %v", statusErr)
		if statusErr := r.updatePhase(ctx, instance, apiv3.CRFailed); statusErr != nil {
			klog.Error(statusErr)
		}
		klog.Errorf("Fail to reconcile %s/%s: %v", instance.Namespace, instance.Name, statusErr)
		return ctrl.Result{}, statusErr
	}

	if statusErr = r.Bootstrap.PropagateDefaultCR(instance); statusErr != nil {
		klog.Error(statusErr)
		return ctrl.Result{}, statusErr
	}

	if statusErr = r.Bootstrap.UpdateResourceLabel(instance); statusErr != nil {
		klog.Error(statusErr)
		return ctrl.Result{}, statusErr
	}

	if statusErr = r.Bootstrap.UpdateManageCertRotationLabel(instance); statusErr != nil {
		klog.Error(statusErr)
		return ctrl.Result{}, statusErr
	}

	// Set Succeeded phase
	if statusErr = r.updatePhase(ctx, instance, apiv3.CRSucceeded); statusErr != nil {
		klog.Error(statusErr)
		return ctrl.Result{}, statusErr
	}

	if optStatusReady, optStatusErr := r.Bootstrap.CheckSubOperatorStatus(instance); optStatusErr != nil {
		klog.Errorf("Failed to check the status of the operators in the OperandRegistry: %v", optStatusErr)
		return ctrl.Result{}, optStatusErr
	} else if !optStatusReady {
		klog.Infof("Operators in the OperandRegistry are not deployed yet, skip operator status update")
	}

	klog.Infof("Finished reconciling CommonService: %s/%s", instance.Namespace, instance.Name)
	return ctrl.Result{}, nil
}

// ReconcileGeneralCR is for setting the OperandConfig
func (r *CommonServiceReconciler) ReconcileGeneralCR(ctx context.Context, instance *apiv3.CommonService) (ctrl.Result, error) {

	if instance.Status.Phase == "" {
		if err := r.updatePhase(ctx, instance, apiv3.CRInitializing); err != nil {
			klog.Error(err)
			return ctrl.Result{}, err
		}
	} else {
		if err := r.updatePhase(ctx, instance, apiv3.CRUpdating); err != nil {
			klog.Error(err)
			return ctrl.Result{}, err
		}
	}

	instance.UpdateNonMasterConfigStatus(&r.Bootstrap.CSData)

	opcon := util.NewUnstructured("operator.ibm.com", "OperandConfig", "v1alpha1")
	opconKey := types.NamespacedName{
		Name:      "common-service",
		Namespace: r.Bootstrap.CSData.ServicesNs,
	}
	if err := r.Reader.Get(ctx, opconKey, opcon); err != nil {
		klog.Errorf("failed to get OperandConfig %s: %v", opconKey.String(), err)
		if err := r.updatePhase(ctx, instance, apiv3.CRFailed); err != nil {
			klog.Error(err)
		}
		klog.Errorf("Fail to reconcile %s/%s: %v", instance.Namespace, instance.Name, err)
		return ctrl.Result{}, err
	}

	cs := util.NewUnstructured("operator.ibm.com", "CommonService", "v3")
	if err := r.Bootstrap.Client.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, cs); err != nil {
		klog.Errorf("Fail to reconcile %s/%s: %v", instance.Namespace, instance.Name, err)
		return ctrl.Result{}, err
	}
	// Generate Issuer and Certificate CR
	if err := r.Bootstrap.DeployCertManagerCR(); err != nil {
		klog.Errorf("Failed to deploy cert manager CRs: %v", err)
		if err := r.updatePhase(ctx, instance, apiv3.CRFailed); err != nil {
			klog.Error(err)
		}
		klog.Errorf("Fail to reconcile %s/%s: %v", instance.Namespace, instance.Name, err)
		return ctrl.Result{}, err
	}

	newConfigs, serviceControllerMapping, err := r.getNewConfigs(cs)
	if err != nil {
		if err := r.updatePhase(ctx, instance, apiv3.CRFailed); err != nil {
			klog.Error(err)
		}
		klog.Errorf("Fail to reconcile %s/%s: %v", instance.Namespace, instance.Name, err)
		return ctrl.Result{}, err
	}

	isEqual, err := r.updateOperandConfig(ctx, newConfigs, serviceControllerMapping)
	if err != nil {
		if err := r.updatePhase(ctx, instance, apiv3.CRFailed); err != nil {
			klog.Error(err)
		}
		klog.Errorf("Fail to reconcile %s/%s: %v", instance.Namespace, instance.Name, err)
		return ctrl.Result{}, err
	}

	// Create Event if there is no update in OperandConfig after applying current CR
	if isEqual {
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Noeffect", fmt.Sprintf("No update, resource sizings in the OperandConfig %s/%s are larger than the profile from CommonService CR %s/%s", r.Bootstrap.CSData.OperatorNs, "common-service", instance.Namespace, instance.Name))
	}

	isEqual, err = r.updateOperatorConfig(ctx, instance.Spec.OperatorConfigs)
	if err != nil {
		if err := r.updatePhase(ctx, instance, apiv3.CRFailed); err != nil {
			klog.Error(err)
		}
		klog.Errorf("Fail to reconcile %s/%s: %v", instance.Namespace, instance.Name, err)
		return ctrl.Result{}, err
	}

	// Wait for Postgres Cluster image to be updated
	if err := r.Bootstrap.UpdatePostgresClusterImage(ctx, instance); err != nil {
		klog.Errorf("Failed to update Postgres Cluster image: %v", err)
		if err := r.updatePhase(ctx, instance, apiv3.CRFailed); err != nil {
			klog.Error(err)
		}
		klog.Errorf("Fail to reconcile %s/%s: %v", instance.Namespace, instance.Name, err)
		return ctrl.Result{}, err
	}

	if err := r.Bootstrap.UpdateResourceLabel(instance); err != nil {
		klog.Error(err)
		return ctrl.Result{}, err
	}

	if err = r.Bootstrap.UpdateManageCertRotationLabel(instance); err != nil {
		klog.Error(err)
		return ctrl.Result{}, err
	}

	// Create Event if there is no update in OperatorConfig after applying current CR
	if isEqual {
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Noeffect", fmt.Sprintf("No update to, replica sizings in the OperatorConfig %s/%s are larger than the profile from CommonService CR %s/%s", r.Bootstrap.CSData.OperatorNs, "test-operator-config", instance.Namespace, instance.Name))
	}

	// Set Ready condition
	instance.SetReadyCondition(constant.KindCR, apiv3.ConditionTypeReady, corev1.ConditionTrue)
	if err := r.Client.Status().Update(ctx, instance); err != nil {
		klog.Warning(err)
		return ctrl.Result{}, err
	}

	if err := r.updatePhase(ctx, instance, apiv3.CRSucceeded); err != nil {
		klog.Errorf("Fail to reconcile %s/%s: %v", instance.Namespace, instance.Name, err)
		return ctrl.Result{}, err
	}

	klog.Infof("Finished reconciling CommonService: %s/%s", instance.Namespace, instance.Name)
	return ctrl.Result{}, nil
}

// ReconileNonConfigurableCR is for setting the cloned Master CR status for advanced topologies
func (r *CommonServiceReconciler) ReconcileNonConfigurableCR(ctx context.Context, instance *apiv3.CommonService) (ctrl.Result, error) {

	if instance.Status.Phase == "" {
		if err := r.updatePhase(ctx, instance, apiv3.CRInitializing); err != nil {
			klog.Error(err)
			return ctrl.Result{}, err
		}
	} else {
		if err := r.updatePhase(ctx, instance, apiv3.CRUpdating); err != nil {
			klog.Error(err)
			return ctrl.Result{}, err
		}
	}

	originalInstance := instance.DeepCopy()
	instance.UpdateNonMasterConfigStatus(&r.Bootstrap.CSData)

	if !reflect.DeepEqual(originalInstance.Status, instance.Status) {
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Noeffect", fmt.Sprintf("No update, this resource is the clone of Common Service CR named %s from namespace %s", constant.MasterCR, r.Bootstrap.CSData.OperatorNs))
		if err := r.Client.Status().Update(ctx, instance); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Set Ready condition
	instance.SetReadyCondition(constant.KindCR, apiv3.ConditionTypeReady, corev1.ConditionTrue)
	if err := r.Client.Status().Update(ctx, instance); err != nil {
		klog.Warning(err)
		return ctrl.Result{}, err
	}

	if err := r.updatePhase(ctx, instance, apiv3.CRSucceeded); err != nil {
		klog.Errorf("Fail to reconcile %s/%s: %v", instance.Namespace, instance.Name, err)
		return ctrl.Result{}, err
	}

	klog.Infof("Finished reconciling CommonService: %s/%s", instance.Namespace, instance.Name)
	return ctrl.Result{}, nil
}

func (r *CommonServiceReconciler) mappingToCsRequestForConfigMaps(ctx context.Context, object client.Object) []reconcile.Request {
	configMap, ok := object.(*corev1.ConfigMap)
	if !ok {
		return nil
	}

	// Check two configmaps: common-service-maps and ibm-cpp-config
	if (configMap.Name == constant.CsMapConfigMap && configMap.Namespace == constant.CsMapConfigMapNs) ||
		(configMap.Name == constant.IBMCPPCONFIG && configMap.Namespace == r.Bootstrap.CSData.ServicesNs) {
		return []reconcile.Request{
			{NamespacedName: types.NamespacedName{
				Name:      constant.MasterCR,
				Namespace: r.Bootstrap.CSData.OperatorNs,
			}},
		}
	}
	return nil
}

func (r *CommonServiceReconciler) mappingToCsRequestForOperandRegistry(ctx context.Context, object client.Object) []reconcile.Request {
	operandRegistry, ok := object.(*odlm.OperandRegistry)
	if !ok {
		// It's not an OperandRegistry, ignore
		return nil
	}
	if operandRegistry.Name == constant.MasterCR && operandRegistry.Namespace == r.Bootstrap.CSData.ServicesNs {
		if isNonNoopOperandReconcile(operandRegistry) {
			// Enqueue a reconciliation request for the corresponding CommonService
			return []reconcile.Request{
				{NamespacedName: types.NamespacedName{Name: constant.MasterCR, Namespace: r.Bootstrap.CSData.OperatorNs}},
			}
		}
	}
	return nil
}

func (r *CommonServiceReconciler) isODLMManagedSubscription(ctx context.Context, object client.Object) []reconcile.Request {
	subscription, ok := object.(*olmv1alpha1.Subscription)
	if !ok {
		// It's not an Subscription, ignore
		return nil
	}
	if subscription.GetLabels()[constant.OpreqLabel] == "true" {
		// Enqueue a reconciliation request for the corresponding CommonService
		return []reconcile.Request{
			{NamespacedName: types.NamespacedName{Name: constant.MasterCR, Namespace: r.Bootstrap.CSData.OperatorNs}},
		}
	}
	return nil

}

// isNonNoopOperandReconcile checks only no-op operators trigger the reconcile
func isNonNoopOperandReconcile(operandRegistry *odlm.OperandRegistry) bool {

	if operandRegistry.Status.OperatorsStatus != nil {
		// List all requested operators
		for operator := range operandRegistry.Status.OperatorsStatus {
			// If there is a requested operator's installMode is "no-op", then skip reconcile
			for _, op := range operandRegistry.Spec.Operators {
				if op.Name == operator && op.InstallMode == "no-op" {
					klog.Infof("The operator %s with 'no-op' installMode is still requested in OperandRegistry, skip reconciliation", operator)
					return false
				}
			}
		}
	}
	return true
}

func (r *CommonServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {

	controller := ctrl.NewControllerManagedBy(mgr).
		// AnnotationChangedPredicate is intended to be used in conjunction with the GenerationChangedPredicate
		For(&apiv3.CommonService{}, builder.WithPredicates(
			predicate.Or(
				predicate.GenerationChangedPredicate{},
				predicate.AnnotationChangedPredicate{},
				predicate.LabelChangedPredicate{}))).
		Watches(
			&corev1.ConfigMap{},
			handler.EnqueueRequestsFromMapFunc(r.mappingToCsRequestForConfigMaps),
			builder.WithPredicates(predicate.Funcs{
				CreateFunc: func(e event.CreateEvent) bool { return true },
				UpdateFunc: func(e event.UpdateEvent) bool { return true },
				DeleteFunc: func(e event.DeleteEvent) bool { return !e.DeleteStateUnknown },
			}))
	if isOpregAPI, err := r.Bootstrap.CheckCRD(constant.OpregAPIGroupVersion, constant.OpregKind); err != nil {
		klog.Errorf("Failed to check if OperandRegistry CRD exists: %v", err)
		return err
	} else if isOpregAPI {
		controller = controller.Watches(
			&odlm.OperandRegistry{},
			handler.EnqueueRequestsFromMapFunc(r.mappingToCsRequestForOperandRegistry),
			builder.WithPredicates(predicate.Funcs{
				UpdateFunc: func(e event.UpdateEvent) bool {
					oldOperandRegistry, ok := e.ObjectOld.(*odlm.OperandRegistry)
					if !ok {
						return false
					}

					newOperandRegistry, ok := e.ObjectNew.(*odlm.OperandRegistry)
					if !ok {
						return false
					}

					// Return true if the length of .status.operatorsStatus array has changed, indicating that a operator has been added or removed
					return len(oldOperandRegistry.Status.OperatorsStatus) != len(newOperandRegistry.Status.OperatorsStatus)
				},
			},
			))
	}
	if isSubscriptionAPI, err := r.Bootstrap.CheckCRD(constant.SubscriptionAPIGroupVersion, constant.SubscriptionKind); err != nil {
		klog.Errorf("Failed to check if Subscription CRD exists: %v", err)
		return err
	} else if isSubscriptionAPI {
		klog.Infof("Subscription CRD exists, start watching Subscription")
		controller = controller.Watches(
			&olmv1alpha1.Subscription{},
			handler.EnqueueRequestsFromMapFunc(r.isODLMManagedSubscription),
			builder.WithPredicates(predicate.Funcs{
				UpdateFunc: func(e event.UpdateEvent) bool {
					oldObject := e.ObjectOld.(*olmv1alpha1.Subscription)
					newObject := e.ObjectNew.(*olmv1alpha1.Subscription)
					return (oldObject.Status.InstalledCSV != "" && newObject.Status.InstalledCSV != "" && oldObject.Status.InstalledCSV != newObject.Status.InstalledCSV)
				},
				DeleteFunc: func(e event.DeleteEvent) bool {
					return !e.DeleteStateUnknown
				},
				CreateFunc: func(e event.CreateEvent) bool {
					return true
				},
			},
			))
	}
	return controller.Complete(r)
}
