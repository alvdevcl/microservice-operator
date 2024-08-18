package controllers

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	examplev1 "github.com/alvdevcl/microservice-operator/api/v1"
)

// MicroserviceReconciler reconciles a Microservice object
type MicroserviceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=example.com,resources=microservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=example.com,resources=microservices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=example.com,resources=microservices/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete

func (r *MicroserviceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the Microservice instance
	microservice := &examplev1.Microservice{}
	err := r.Get(ctx, req.NamespacedName, microservice)
	if err != nil {
		if errors.IsNotFound(err) {
			// Microservice resource not found, could have been deleted after reconcile request.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	// Define a new Deployment object
	deployment := r.deploymentForMicroservice(microservice)
	// Check if the Deployment already exists
	foundDeployment := &appsv1.Deployment{}
	err = r.Get(ctx, req.NamespacedName, foundDeployment)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
		err = r.Create(ctx, deployment)
		if err != nil {
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}

	// Define a new Service object
	service := r.serviceForMicroservice(microservice)
	// Check if the Service already exists
	foundService := &corev1.Service{}
	err = r.Get(ctx, req.NamespacedName, foundService)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
		err = r.Create(ctx, service)
		if err != nil {
			return ctrl.Result{}, err
		}
		// Service created successfully - return and requeue
		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}

	// Define a new Ingress object
	ingress := r.ingressForMicroservice(microservice)
	// Check if the Ingress already exists
	foundIngress := &networkingv1.Ingress{}
	err = r.Get(ctx, req.NamespacedName, foundIngress)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Ingress", "Ingress.Namespace", ingress.Namespace, "Ingress.Name", ingress.Name)
		err = r.Create(ctx, ingress)
		if err != nil {
			return ctrl.Result{}, err
		}
		// Ingress created successfully - return and requeue
		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *MicroserviceReconciler) deploymentForMicroservice(m *examplev1.Microservice) *appsv1.Deployment {
	replicas := int32(1)
	if m.Spec.Replicas != nil {
		replicas = *m.Spec.Replicas
	}
	labels := map[string]string{"app": m.Spec.Name}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Spec.Name,
			Namespace: m.Namespace,
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
						Name:  m.Spec.Name,
						Image: m.Spec.Image,
						Ports: []corev1.ContainerPort{{
							ContainerPort: m.Spec.Port,
							Name:          "http",
						}},
					}},
				},
			},
		},
	}
}

func (r *MicroserviceReconciler) serviceForMicroservice(m *examplev1.Microservice) *corev1.Service {
	labels := map[string]string{"app": m.Spec.Name}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Spec.Name,
			Namespace: m.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Port:     80,
					TargetPort: intstr.FromInt(int(m.Spec.Port)),
				},
			},
		},
	}
}

func (r *MicroserviceReconciler) ingressForMicroservice(m *examplev1.Microservice) *networkingv1.Ingress {
	// Remove the labels variable as it's unused
	return &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Spec.Name,
			Namespace: m.Namespace,
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: m.Spec.IngressHost,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path: "/",
									// Use this instead of PathType: (*networkingv1.PathType)(intstr.FromString("Prefix"))
									PathType: func() *networkingv1.PathType {
										pathType := networkingv1.PathTypePrefix
										return &pathType
									}(),
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: m.Spec.Name,
											Port: networkingv1.ServiceBackendPort{
												Number: 80,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *MicroserviceReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&examplev1.Microservice{}).
        Owns(&appsv1.Deployment{}).
        Owns(&corev1.Service{}).
        Owns(&networkingv1.Ingress{}).
        Complete(r)
}