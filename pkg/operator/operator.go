package operator

import (
	"fmt"

	"github.com/blang/semver"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers/core/v1"
	coreclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	operatorsv1alpha1 "github.com/openshift/api/operator/v1alpha1"
	"github.com/openshift/library-go/pkg/operator/resource/resourceapply"
	"github.com/openshift/library-go/pkg/operator/versioning"

	"github.com/enj/example-operator/pkg/controller"
)

const (
	TargetNamespace = "example-operator"

	workQueueKey = "key"
)

func NewExampleOperator(coreInformers v1.Interface, secretsClient coreclientv1.SecretsGetter, configMapsClient coreclientv1.ConfigMapsGetter) *ExampleOperator {
	c := &ExampleOperator{
		secretsClient:    secretsClient,
		configMapsClient: configMapsClient,
	}

	// TODO move this logic up one level
	secretsInformer := coreInformers.Secrets().Informer()
	configMapsInformer := coreInformers.ConfigMaps().Informer()

	embeddedController, queue := controller.New("ExampleOperator", c.sync, secretsInformer.HasSynced, configMapsInformer.HasSynced)

	c.Controller = embeddedController

	secretsInformer.AddEventHandler(eventHandler(queue))
	configMapsInformer.AddEventHandler(eventHandler(queue))

	return c
}

// eventHandler queues the operator to check spec and status
// TODO add filtering and more nuanced logic
func eventHandler(queue workqueue.RateLimitingInterface) cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { queue.Add(workQueueKey) },
		UpdateFunc: func(old, new interface{}) { queue.Add(workQueueKey) },
		DeleteFunc: func(obj interface{}) { queue.Add(workQueueKey) },
	}
}

type ExampleOperator struct {
	secretsClient    coreclientv1.SecretsGetter
	configMapsClient coreclientv1.ConfigMapsGetter

	*controller.Controller
}

func (c *ExampleOperator) sync() error {
	config, err := c.configMapsClient.ConfigMaps(TargetNamespace).Get("instance", metav1.GetOptions{})
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
		return c.secretsClient.Secrets(TargetNamespace).Delete(secretName, nil)

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
		_, _, err := resourceapply.ApplySecret(c.secretsClient, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: TargetNamespace,
			},
			Data: map[string][]byte{
				secretData: []byte("007"),
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

	_, _, err = resourceapply.ApplyConfigMap(c.configMapsClient, outConfig)
	errs = append(errs, err)

	return utilerrors.NewAggregate(errs)
}
