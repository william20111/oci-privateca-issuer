/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	ocicav1alpha1 "github.com/william20111/oci-privateca-issuer/pkg/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

// OCICAClusterIssuerReconciler reconciles a OCICAClusterIssuer object
type OCICAClusterIssuerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=ocica.cert-manager.io,resources=ocicaclusterissuers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ocica.cert-manager.io,resources=ocicaclusterissuers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ocica.cert-manager.io,resources=ocicaclusterissuers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the OCICAClusterIssuer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *OCICAClusterIssuerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	iss := new(ocicav1alpha1.OCICAClusterIssuer)
	err := r.Client.Get(ctx, req.NamespacedName, iss)
	if err != nil {
		logger.Error(err, "failed fetch oci issuer")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	err = validateIssuer(iss.Spec)
	if err != nil {
		logger.Error(err, "failed to validate resource spec")
		return ctrl.Result{}, err
	}

	return reconcile.Result{}, r.setStatus(ctx, iss, ocicav1alpha1.ConditionTrue, "Verified", "OriginIssuer verified and ready to sign certificates")
}

// SetupWithManager sets up the controller with the Manager.
func (r *OCICAClusterIssuerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ocicav1alpha1.OCICAClusterIssuer{}).
		Complete(r)
}

// setStatus is a function to set the issuer status
func (r *OCICAClusterIssuerReconciler) setStatus(ctx context.Context, iss *ocicav1alpha1.OCICAClusterIssuer, status metav1.ConditionStatus, reason, message string) error {
	now := metav1.NewTime(time.Now())
	c := metav1.Condition{
		Type:               string(ocicav1alpha1.ConditionReady),
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: now,
	}

	for i, condition := range iss.Status.Conditions {
		if condition.Type != string(ocicav1alpha1.ConditionReady) {
			continue
		}

		if condition.Status == status {
			c.LastTransitionTime = condition.LastTransitionTime
		}
		iss.Status.Conditions[i] = c
	}

	iss.Status.Conditions = append(iss.Status.Conditions, c)
	return r.Client.Status().Update(ctx, iss)
}

func validateIssuer(spec ocicav1alpha1.OCICAClusterIssuerSpec) error {
	switch {
	case spec.SecretRef.User != "":
		return fmt.Errorf("user cant be empty in secret config")
	case spec.SecretRef.Name != "":
		return fmt.Errorf("name cant be empty in secret config")
	case spec.SecretRef.Region != "":
		return fmt.Errorf("region cant be empty in secret config")
	case spec.SecretRef.FingerPrint != "":
		return fmt.Errorf("fingerprint cant be empty in secret config")
	case spec.SecretRef.PrivateKey != "":
		return fmt.Errorf("private key cant be empty in secret config")
	case spec.SecretRef.PrivateKeyPassphrase != "":
		return fmt.Errorf("passphrase cant be empty in secret config")
	}
	return nil
}
