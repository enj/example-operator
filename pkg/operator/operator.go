package operator

import (
	"fmt"
	"time"

	"github.com/blang/semver"
	"github.com/golang/glog"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers/core/v1"
	coreclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	operatorsv1alpha1 "github.com/openshift/api/operator/v1alpha1"
	"github.com/openshift/library-go/pkg/operator/resource/resourceapply"
	"github.com/openshift/library-go/pkg/operator/versioning"
)

const (
	targetNamespaceName = "example-operator"
	workQueueKey        = "key"
)

type ExampleOperator struct {
	secret    coreclientv1.SecretsGetter
	configMap coreclientv1.ConfigMapsGetter

	// allows for unit testing
	syncHandler func() error

	// queue only ever has one item, but it has nice error handling backoff/retry semantics
	queue workqueue.RateLimitingInterface
}

func NewExampleOperator(
	informers v1.Interface,
	secret coreclientv1.SecretsGetter,
	configMap coreclientv1.ConfigMapsGetter,
) *ExampleOperator {
	c := &ExampleOperator{
		secret:    secret,
		configMap: configMap,

		queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ExampleOperator"),
	}

	c.syncHandler = c.sync

	informers.Secrets().Informer().AddEventHandler(c.eventHandler())
	informers.ConfigMaps().Informer().AddEventHandler(c.eventHandler())

	return c
}

func (c ExampleOperator) sync() error {
	config, err := c.configMap.ConfigMaps(targetNamespaceName).Get("instance", metav1.GetOptions{})
	if err != nil {
		return err
	}

	// these are my pretend spec/status fields
	d := config.Data

	secretName := d["name"]

	state := operatorsv1alpha1.ManagementState(d["state"])

	switch state {
	case operatorsv1alpha1.Managed:
		// handled below

	case operatorsv1alpha1.Unmanaged:
		return nil

	case operatorsv1alpha1.Removed:
		return c.secret.Secrets(targetNamespaceName).Delete(secretName, nil)

	default:
		// TODO should update status
		return fmt.Errorf("unknown state: %v", state)
	}

	startVersion := d["current_version"]
	endVersion := d["desired_version"]

	var currentActualVerion *semver.Version

	if len(startVersion) != 0 {
		ver, err := semver.Parse(startVersion)
		if err != nil {
			utilruntime.HandleError(err)
		} else {
			currentActualVerion = &ver
		}
	}
	desiredVersion, err := semver.Parse(endVersion)
	if err != nil {
		// TODO report failing status, we may actually attempt to do this in the "normal" error handling
		return err
	}

	v310_00_to_unknown := versioning.NewRangeOrDie("3.10.0", "3.10.1")

	outConfig := config.DeepCopy()
	var errs []error

	switch {
	case v310_00_to_unknown.BetweenOrEmpty(currentActualVerion) && v310_00_to_unknown.Between(&desiredVersion):
		secretData := d["data"]
		_, _, err := resourceapply.ApplySecret(c.secret, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: targetNamespaceName,
			},
			StringData: map[string]string{
				"data": secretData,
			},
		})
		errs = append(errs, err)

		if err == nil { // this needs work
			outConfig.Data["summary"] = "sync-[3.10.0,3.10.1)"
			outConfig.Data["current_version"] = desiredVersion.String()
		}

	default:
		outConfig.Data["summary"] = "unrecognized"
	}

	_, _, err = resourceapply.ApplyConfigMap(c.configMap, outConfig)
	errs = append(errs, err)

	return utilerrors.NewAggregate(errs)
}

// Run starts the serviceCertSigner and blocks until stopCh is closed.
func (c *ExampleOperator) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	glog.Infof("Starting ExampleOperator")
	defer glog.Infof("Shutting down ExampleOperator")

	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
}

func (c *ExampleOperator) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *ExampleOperator) processNextWorkItem() bool {
	dsKey, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(dsKey)

	err := c.syncHandler()
	if err == nil {
		c.queue.Forget(dsKey)
		return true
	}

	utilruntime.HandleError(fmt.Errorf("%v failed with : %v", dsKey, err))
	c.queue.AddRateLimited(dsKey)

	return true
}

// eventHandler queues the operator to check spec and status
// TODO add filtering
func (c *ExampleOperator) eventHandler() cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { c.queue.Add(workQueueKey) },
		UpdateFunc: func(old, new interface{}) { c.queue.Add(workQueueKey) },
		DeleteFunc: func(obj interface{}) { c.queue.Add(workQueueKey) },
	}
}
