package main

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	corev1informer "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1"
	corev1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	samplev1alpha1 "github.com/raihankhan/httpApiServer-controller/pkg/apis/raihankhan.github.io/v1alpha1"
	clientset "github.com/raihankhan/httpApiServer-controller/pkg/client/clientset/versioned"
	informers "github.com/raihankhan/httpApiServer-controller/pkg/client/informers/externalversions/raihankhan.github.io/v1alpha1"
	listers "github.com/raihankhan/httpApiServer-controller/pkg/client/listers/raihankhan.github.io/v1alpha1"
)

const controllerAgentName = "httpApiServer-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when a Foo is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a Foo fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by ApiServer"
	// MessageResourceSynced is the message used for an Event fired when a Foo
	// is synced successfully
	MessageResourceSynced = "Apiserver synced successfully"
)

// Controller is the controller implementation for Apiserver resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// sampleclientset is a clientset for our own API group
	sampleclientset clientset.Interface

	deploymentsLister appslisters.DeploymentLister
	deploymentsSynced cache.InformerSynced
	apiServerLister   listers.ApiserverLister
	apiServersSynced  cache.InformerSynced
	serviceLister     corev1lister.ServiceLister
	serviceSynced     cache.InformerSynced


	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
}

// NewController returns a new sample controller
func NewController(
	kubeclientset kubernetes.Interface,
	sampleclientset clientset.Interface,
	deploymentInformer appsinformers.DeploymentInformer,
	serviceInformer corev1informer.ServiceInformer,
	apiServerInformer informers.ApiserverInformer) *Controller {

	controller := &Controller{
		kubeclientset:     kubeclientset,
		sampleclientset:   sampleclientset,
		deploymentsLister: deploymentInformer.Lister(),
		deploymentsSynced: deploymentInformer.Informer().HasSynced,
		apiServerLister:   apiServerInformer.Lister(),
		apiServersSynced:  apiServerInformer.Informer().HasSynced,
		serviceSynced:     serviceInformer.Informer().HasSynced,
		serviceLister:     serviceInformer.Lister(),

		workqueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "apiservers"),
	}

	klog.Info("Setting up event handlers")
	// Set up an event handler for when Foo resources change

	apiServerInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueApiServer,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueApiServer(new)
		},
	})
	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(workers int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting Apiserver controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.deploymentsSynced, c.serviceSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	// Launch two workers to process Foo resources
	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}



// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Foo resource
// with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)

	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the Apiserver resource with this namespace/name
	apiServer, err := c.apiServerLister.Apiservers(namespace).Get(name)

	if err != nil {
		// The Apiserver resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("Apiserver '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	//deployment
	//
	//
	deploymentName := apiServer.Spec.DeploymentName
	if deploymentName == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		utilruntime.HandleError(fmt.Errorf("%s: deployment name must be specified", key))
		return nil
	}

	//Get the deployment with the name specified in Apiserver.spec
	deployment, err := c.deploymentsLister.Deployments(apiServer.Namespace).Get(deploymentName)
	//fmt.Println(err.Error())
	//If the resource doesn't exist, we'll create it
	if err != nil && strings.Contains(err.Error(), "not found") {
		fmt.Println(deploymentName)
		//spew.Dump(apiServer.Spec)
		deployment, err = c.kubeclientset.AppsV1().Deployments(apiServer.Namespace).Create(context.TODO(), newDeployment(apiServer), metav1.CreateOptions{})
		fmt.Println(err.Error())
		//
	}
	fmt.Println("came here")
	fmt.Println(deployment.Name + "---")

	//Service
	//
	//
	serviceName := apiServer.Spec.ServiceName + "-svc"
	if serviceName == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		utilruntime.HandleError(fmt.Errorf("%s: service name must be specified", key))
		return nil
	}

	// Get the deployment with the name specified in Apiserver.spec
	svc, err := c.serviceLister.Services(apiServer.Namespace).Get(serviceName)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		svc, err = c.kubeclientset.CoreV1().Services(apiServer.Namespace).Create(context.TODO(), newService(apiServer), metav1.CreateOptions{})
	}

	fmt.Println(svc.Name)

	//NodePort
	//
	//
	NodePortName := apiServer.Spec.NodePortName + "-np"
	if serviceName == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		utilruntime.HandleError(fmt.Errorf("%s: NodePortName name must be specified", key))
		return nil
	}

	// Get the deployment with the name specified in Apiserver.spec
	np, err := c.serviceLister.Services(apiServer.Namespace).Get(NodePortName)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		np, err = c.kubeclientset.CoreV1().Services(apiServer.Namespace).Create(context.TODO(), newNodePort(apiServer), metav1.CreateOptions{})
	}

	fmt.Println(np.Name)

	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// Finally, we update the status block of the Foo resource to reflect the
	// current state of the world
	err = c.updateServerStatus(apiServer, deployment)
	if err != nil {
		return err
	}

	return nil
}

func (c *Controller) updateServerStatus(server *samplev1alpha1.Apiserver, deployment *appsv1.Deployment) error {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	serverCopy := server.DeepCopy()
	serverCopy.Status.AvailableReplicas = deployment.Status.AvailableReplicas
	// If the CustomResourceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the Foo resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.
	_, err := c.sampleclientset.RaihankhanV1alpha1().Apiservers(server.Namespace).Update(context.TODO(), serverCopy, metav1.UpdateOptions{})
	return err
}

// enqueueApiserver takes a Foo resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Foo.
func (c *Controller) enqueueApiServer(obj interface{}) {
	var key string
	var err error
	klog.Info("Enqueueing Apiservers. . . ")
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}

// newDeployment creates a new Deployment for a Apiserver resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the ApiServer resource that 'owns' it.
func newDeployment(apiServer *samplev1alpha1.Apiserver) *appsv1.Deployment {
	labels := map[string]string{
		"app":        "apiserver",
		"controller": apiServer.Name,
	}

	port1 := corev1.ContainerPort{
		Name:          "http",
		Protocol:      corev1.ProtocolTCP,
		ContainerPort: 8080,
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      apiServer.Spec.DeploymentName,
			Namespace: apiServer.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(apiServer, samplev1alpha1.SchemeGroupVersion.WithKind("Apiserver")),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: apiServer.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							//Name:  "httpapiserver",
							Image: "raihankhanraka/ecommerce-api:v1.1",
							Ports: []corev1.ContainerPort{
								port1,
							},
						},
					},
				},
			},
		},
	}
}

// newService creates a new Deployment for a Apiserver resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the ApiServer resource that 'owns' it.
func newService(apiServer *samplev1alpha1.Apiserver) *corev1.Service {
	selectors := map[string]string{
		"app":        "apiserver",
		"controller": apiServer.Name,
	}

	port1 := corev1.ServicePort{
		Protocol:   corev1.ProtocolTCP,
		TargetPort: intstr.IntOrString{IntVal: 8080},
		Port:       8081,
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      apiServer.Spec.ServiceName,
			Namespace: apiServer.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(apiServer, samplev1alpha1.SchemeGroupVersion.WithKind("Apiserver")),
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				port1,
			},
			Selector: selectors,
		},
	}
}

// newNodePort creates a new Deployment for a Apiserver resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the ApiServer resource that 'owns' it.
func newNodePort(apiServer *samplev1alpha1.Apiserver) *corev1.Service {
	selectors := map[string]string{
		"app":        "apiserver",
		"controller": apiServer.Name,
	}

	port1 := corev1.ServicePort{
		Protocol:   corev1.ProtocolTCP,
		TargetPort: intstr.IntOrString{IntVal: 8081},
		Port:       8081,
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      apiServer.Spec.NodePortName + "-svc",
			Namespace: apiServer.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(apiServer, samplev1alpha1.SchemeGroupVersion.WithKind("Apiserver")),
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: selectors,
			Type:     corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{
				port1,
			},
		},
	}
}
