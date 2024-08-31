import (
	...
	microservicev1alpha1 "github.com/alvdevcl/microservice-operator/api/v1alpha1"
	"github.com/alvdevcl/microservice-operator/controllers"
)

func main() {
	...
	if err = (&controllers.CoreUIReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CoreUI")
		os.Exit(1)
	}

	// Repeat for each microservice
	if err = (&controllers.AuthenticationServiceReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AuthenticationService")
		os.Exit(1)
	}
	// ...
}