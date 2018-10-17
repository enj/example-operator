package starter

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/enj/example-operator/pkg/operator"
)

func RunOperator(clientConfig *rest.Config, stopCh <-chan struct{}) error {
	kubeClient, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return err
	}

	kubeInformersNamespaced := informers.NewSharedInformerFactoryWithOptions(kubeClient, 10*time.Minute,
		informers.WithNamespace(operator.TargetNamespace),
		informers.WithTweakListOptions(func(options *v1.ListOptions) {
			options.FieldSelector = fields.OneTermEqualSelector("metadata.name", operator.ResourceName).String()
		}),
	)

	coreInformers := kubeInformersNamespaced.Core().V1()
	secretsInformer := coreInformers.Secrets()
	configMapsInformer := coreInformers.ConfigMaps()

	exampleOperator := operator.NewExampleOperator(
		configMapsInformer,
		secretsInformer,
		kubeClient.CoreV1(),
		kubeClient.CoreV1(),
	)

	kubeInformersNamespaced.Start(stopCh)

	exampleOperator.Run(stopCh)

	return fmt.Errorf("stopped")
}
