/*

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

package networking

import (
	"context"
	"fmt"

	"code.cloudfoundry.org/cf-k8s-networking/routecontroller/resourcebuilders"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	istionetworkingv1alpha3 "code.cloudfoundry.org/cf-k8s-networking/routecontroller/apis/istio/networking/v1alpha3"
	networkingv1alpha1 "code.cloudfoundry.org/cf-k8s-networking/routecontroller/apis/networking/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// RouteReconciler reconciles a Route object
type RouteReconciler struct {
	client.Client
	Log          logr.Logger
	Scheme       *runtime.Scheme
	IstioGateway string
}

// +kubebuilder:rbac:groups=networking.cloudfoundry.org,resources=routes,verbs=get;list;watch
// +kubebuilder:rbac:groups=networking.cloudfoundry.org,resources=routes/status,verbs=get;update;patch

func (r *RouteReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("route", req.NamespacedName)

	// your logic goes here
	routes := &networkingv1alpha1.RouteList{}

	// TODO: only act on changes to routes? consider doing this in the update story

	// watch finds a new route or change to a route
	// find all routes that share that fqdn and reconcile the single Virtual Service for that fqdn
	// reconcile the many Services for the route that was created/changed

	err := r.List(ctx, routes)
	if err != nil {
		log.Error(err, "failed to list routes")
	}

	vsb := resourcebuilders.VirtualServiceBuilder{IstioGateways: []string{r.IstioGateway}}
	sb := resourcebuilders.ServiceBuilder{}

	virtualservices := vsb.Build(routes)
	services := sb.Build(routes)

	for _, desiredService := range services {
		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      desiredService.ObjectMeta.Name,
				Namespace: desiredService.ObjectMeta.Namespace,
			},
		}
		mutateFn := sb.BuildMutateFunction(service, &desiredService)
		result, err := controllerutil.CreateOrUpdate(ctx, r.Client, service, mutateFn)
		if err != nil {
			return ctrl.Result{}, err
		}
		log.Info(fmt.Sprintf("Service %s/%s has been %s", service.Namespace, service.Name, result))
	}

	for _, desiredVirtualService := range virtualservices {
		virtualService := &istionetworkingv1alpha3.VirtualService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      desiredVirtualService.ObjectMeta.Name,
				Namespace: desiredVirtualService.ObjectMeta.Namespace,
			},
		}
		mutateFn := vsb.BuildMutateFunction(virtualService, &desiredVirtualService)
		result, err := controllerutil.CreateOrUpdate(ctx, r.Client, virtualService, mutateFn)
		if err != nil {
			return ctrl.Result{}, err
		}
		log.Info(fmt.Sprintf("VirtualService %s/%s has been %s", virtualService.Namespace, virtualService.Name, result))
	}

	return ctrl.Result{}, nil
}

func (r *RouteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkingv1alpha1.Route{}).
		Complete(r)
}
