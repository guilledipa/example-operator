package controllers

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	appsv1alpha1 "github.com/guilledipa/example-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ExampleReconciler reconciles a Example object
type ExampleReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps.example.com,resources=examples,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.example.com,resources=examples/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.example.com,resources=examples/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete

func (r *ExampleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling Example", "namespace", req.Namespace, "name", req.Name)

	// Fetch the Example instance
	example := &appsv1alpha1.Example{}
	err := r.Get(ctx, req.NamespacedName, example)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Return and don't requeue
			logger.Info("Example resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		logger.Error(err, "Failed to get Example")
		return ctrl.Result{}, err
	}

	// Check if the ConfigMap already exists, if not create a new one
	configMap := &corev1.ConfigMap{}
	err = r.Get(ctx, client.ObjectKey{Name: example.Name, Namespace: example.Namespace}, configMap)
	if err != nil && errors.IsNotFound(err) {
		// Define a new configmap
		cm := r.configMapForExample(example)
		logger.Info("Creating a new ConfigMap", "Namespace", cm.Namespace, "Name", cm.Name)
		err = r.Create(ctx, cm)
		if err != nil {
			logger.Error(err, "Failed to create new ConfigMap", "Namespace", cm.Namespace, "Name", cm.Name)
			return ctrl.Result{}, err
		}
		// ConfigMap created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		logger.Error(err, "Failed to get ConfigMap")
		return ctrl.Result{}, err
	}

	// Update the ConfigMap if needed
	if configMap.Data["message"] != example.Spec.Message {
		configMap.Data["message"] = example.Spec.Message
		err = r.Update(ctx, configMap)
		if err != nil {
			logger.Error(err, "Failed to update ConfigMap", "Namespace", configMap.Namespace, "Name", configMap.Name)
			return ctrl.Result{}, err
		}
	}

	// Update status condition
	meta.SetStatusCondition(&example.Status.Conditions, metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		Reason:             "ConfigMapCreated",
		Message:            "ConfigMap was created successfully",
		LastTransitionTime: metav1.NewTime(time.Now()),
	})

	if err := r.Status().Update(ctx, example); err != nil {
		logger.Error(err, "Failed to update Example status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// configMapForExample returns a ConfigMap object for an Example resource
func (r *ExampleReconciler) configMapForExample(m *appsv1alpha1.Example) *corev1.ConfigMap {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
			Labels: map[string]string{
				"app": "example-operator",
			},
		},
		Data: map[string]string{
			"message": m.Spec.Message,
		},
	}
	// Set the owner reference
	ctrl.SetControllerReference(m, cm, r.Scheme)
	return cm
}

// SetupWithManager sets up the controller with the Manager.
func (r *ExampleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.Example{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}
