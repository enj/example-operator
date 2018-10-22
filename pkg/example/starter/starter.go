package starter

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/enj/example-operator/pkg/example/operator"
	"github.com/enj/example-operator/pkg/generated/clientset/versioned"
	"github.com/enj/example-operator/pkg/generated/informers/externalversions"
)

func RunOperator(clientConfig *rest.Config, stopCh <-chan struct{}) error {
	kubeClient, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return err
	}

	exampleOperatorClient, err := versioned.NewForConfig(clientConfig)
	if err != nil {
		return err
	}

	const resync = 10 * time.Minute

	// only watch a specific resource name
	tweakListOptions := func(options *v1.ListOptions) {
		options.FieldSelector = fields.OneTermEqualSelector("metadata.name", operator.ResourceName).String()
	}

	kubeInformersNamespaced := informers.NewSharedInformerFactoryWithOptions(kubeClient, resync,
		informers.WithNamespace(operator.TargetNamespace),
		informers.WithTweakListOptions(tweakListOptions),
	)

	exampleOperatorInformers := externalversions.NewSharedInformerFactoryWithOptions(exampleOperatorClient, resync,
		externalversions.WithNamespace(operator.TargetNamespace),
		externalversions.WithTweakListOptions(tweakListOptions),
	)

	exampleOperator := operator.NewExampleOperator(
		exampleOperatorInformers.Exampleoperator().V1alpha1().ExampleOperators(),
		kubeInformersNamespaced.Core().V1().Secrets(),
		exampleOperatorClient.ExampleoperatorV1alpha1().ExampleOperators(operator.TargetNamespace),
		kubeClient.CoreV1(),
	)

	kubeInformersNamespaced.Start(stopCh)
	exampleOperatorInformers.Start(stopCh)

	go exampleOperator.Run(stopCh)

	<-stopCh

	return fmt.Errorf("stopped")
}
