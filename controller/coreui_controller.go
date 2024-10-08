package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	microservicev1alpha1 "github.com/alvdevcl/microservice-operator/api/v1alpha1"
)

// CoreUIReconciler reconciles a CoreUI object
type CoreUIReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=microservice.example.com,resources=coreuis,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=microservice.example.com,resources=coreuis/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=microservice.example.com,resources=coreuis/finalizers,verbs=update

func (r *CoreUIReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the CoreUI instance
	coreui := &microservicev1alpha1.CoreUI{}
	err := r.Get(ctx, req.NamespacedName, coreui)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return. Created objects are automatically garbage collected.
			log.Info("CoreUI resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get CoreUI")
		return ctrl.Result{}, err
	}

	// Define your business logic here, like creating Deployment, Service, Ingress, etc.

	// Example of creating a Deployment for CoreUI (This code should be in a separate function ideally)
	deployment := r.createDeployment(coreui)

	// Check if the deployment already exists, if not create a new one
	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
		err = r.Create(ctx, deployment)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
			return ctrl.Result{}, err
		}
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// Update the CoreUI status (if necessary)
	// coreui.Status.XYZ = "new-status"
	// err = r.Status().Update(ctx, coreui)
	// if err != nil {
	//    log.Error(err, "Failed to update CoreUI status")
	//    return ctrl.Result{}, err
	// }

	return ctrl.Result{}, nil
}

func (r *CoreUIReconciler) createDeployment(coreui *microservicev1alpha1.CoreUI) *appsv1.Deployment {
	labels := map[string]string{"app": "coreui"}
	replicas := coreui.Spec.Replicas

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      coreui.Name,
			Namespace: coreui.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: coreui.Spec.Image,
						Name:  "coreui",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 80,
						}},
					}},
				},
			},
		},
	}

	// Set CoreUI instance as the owner and controller
	ctrl.SetControllerReference(coreui, deployment, r.Scheme)
	return deployment
}

func (r *CoreUIReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&microservicev1alpha1.CoreUI{}).
		Complete(r)
}