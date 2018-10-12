package operator

import (
	"fmt"
	"time"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func RunOperator(clientConfig *rest.Config, stopCh <-chan struct{}) error {
	kubeClient, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		panic(err)
	}

	kubeInformersNamespaced := informers.NewSharedInformerFactoryWithOptions(kubeClient, 10*time.Minute, informers.WithNamespace(targetNamespaceName))

	operator := NewExampleOperator(
		kubeInformersNamespaced.Core().V1(),
		kubeClient.CoreV1(),
		kubeClient.CoreV1(),
	)

	kubeInformersNamespaced.Start(stopCh)

	operator.Run(5, stopCh)
	return fmt.Errorf("stopped")
}
