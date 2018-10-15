package starter

import (
	"fmt"
	"time"

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

	kubeInformersNamespaced := informers.NewSharedInformerFactoryWithOptions(kubeClient, 10*time.Minute, informers.WithNamespace(operator.TargetNamespace))

	operator := operator.NewExampleOperator(
		kubeInformersNamespaced.Core().V1(),
		kubeClient.CoreV1(),
		kubeClient.CoreV1(),
	)

	kubeInformersNamespaced.Start(stopCh)

	operator.Run(1, stopCh) // only start one worker because we only have one key name in our queue

	return fmt.Errorf("stopped")
}
