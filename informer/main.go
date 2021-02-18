package main

import (
	"context"
	"os/signal"
	"os/user"
	"path/filepath"
	"sync"
	"syscall"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func main() {
	u, err := user.Current()
	if err != nil {
		panic(err.Error())
	}
	kubeconfig := filepath.Join(u.HomeDir, ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	informer := cache.NewSharedIndexInformer(
		cache.NewListWatchFromClient(
			clientset.BatchV1().RESTClient(),
			"jobs",
			corev1.NamespaceDefault,
			fields.Everything(),
		),
		// runtime.Object interaceを満たしているのはポインタであることに注意
		&batchv1.Job{},
		// resyncはしない
		0,
		cache.Indexers{},
	)
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			println("Add")
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			println("Update")
		},
		DeleteFunc: func(obj interface{}) {
			println("Delete")
		},
	})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	stopCh := make(chan struct{}, 1)

	// FYI: k8s.io/apimachinery/pkg/util/wait に同様のヘルパーがある
	var wg sync.WaitGroup
	// コンポーネント毎にwg.Add(1)
	wg.Add(1)
	go func() {
		defer wg.Done()
		informer.Run(stopCh)
	}()

	select {
	case <-ctx.Done():
		// 終了処理
		close(stopCh)
		wg.Wait()
	}
}
