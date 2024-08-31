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

// AuthenticationServiceReconciler reconciles a AuthenticationService object
type AuthenticationServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=microservice.example.com,resources=AuthenticationServices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=microservice.example.com,resources=AuthenticationServices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=microservice.example.com,resources=AuthenticationServices/finalizers,verbs=update

func (r *AuthenticationServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the AuthenticationService instance
	AuthenticationService := &microservicev1alpha1.AuthenticationService{}
	err := r.Get(ctx, req.NamespacedName, AuthenticationService)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return. Created objects are automatically garbage collected.
			log.Info("AuthenticationService resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get AuthenticationService")
		return ctrl.Result{}, err
	}

	// Define your business logic here, like creating Deployment, Service, Ingress, etc.

	// Example of creating a Deployment for AuthenticationService (This code should be in a separate function ideally)
	deployment := r.createDeployment(AuthenticationService)

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

	// Update the AuthenticationService status (if necessary)
	// AuthenticationService.Status.XYZ = "new-status"
	// err = r.Status().Update(ctx, AuthenticationService)
	// if err != nil {
	//    log.Error(err, "Failed to update AuthenticationService status")
	//    return ctrl.Result{}, err
	// }

	return ctrl.Result{}, nil
}

func (r *AuthenticationServiceReconciler) createDeployment(AuthenticationService *microservicev1alpha1.AuthenticationService) *appsv1.Deployment {
	labels := map[string]string{"app": "AuthenticationService"}
	replicas := AuthenticationService.Spec.Replicas

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      AuthenticationService.Name,
			Namespace: AuthenticationService.Namespace,
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
						Image: AuthenticationService.Spec.Image,
						Name:  "authentication-service",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 80,
						}},
					}},
				},
			},
		},
	}

	// Set AuthenticationService instance as the owner and controller
	ctrl.SetControllerReference(AuthenticationService, deployment, r.Scheme)
	return deployment
}

func (r *AuthenticationServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&microservicev1alpha1.AuthenticationService{}).
		Complete(r)
}