package starter

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/enj/example-operator/pkg/generated/clientset/versioned"
	"github.com/enj/example-operator/pkg/generated/informers/externalversions"
	"github.com/enj/example-operator/pkg/operator"
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

	kubeInformersNamespaced := informers.NewSharedInformerFactoryWithOptions(kubeClient, 10*time.Minute,
		informers.WithNamespace(operator.TargetNamespace),
		informers.WithTweakListOptions(func(options *v1.ListOptions) {
			options.FieldSelector = fields.OneTermEqualSelector("metadata.name", operator.ResourceName).String()
		}),
	)

	exampleOperatorInformers := externalversions.NewSharedInformerFactoryWithOptions(exampleOperatorClient, 10*time.Minute,
		externalversions.WithNamespace(operator.TargetNamespace),
		externalversions.WithTweakListOptions(func(options *v1.ListOptions) {
			options.FieldSelector = fields.OneTermEqualSelector("metadata.name", operator.ResourceName).String()
		}),
	)

	exampleOperator := operator.NewExampleOperator(
		exampleOperatorInformers.Exampleoperator().V1alpha1().ExampleOperators(),
		kubeInformersNamespaced.Core().V1().Secrets(),
		exampleOperatorClient.ExampleoperatorV1alpha1().ExampleOperators(operator.TargetNamespace),
		kubeClient.CoreV1(),
	)

	kubeInformersNamespaced.Start(stopCh)
	exampleOperatorInformers.Start(stopCh)

	exampleOperator.Run(stopCh)

	return fmt.Errorf("stopped")
}
