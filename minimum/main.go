package main

import (
	"context"
	"fmt"
	"os/user"
	"path/filepath"
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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

	for {
		pods, err := clientset.CoreV1().Pods(apiv1.NamespaceDefault).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		for _, pod := range pods.Items {
			fmt.Printf("%v/%v\n", pod.ObjectMeta.Namespace, pod.ObjectMeta.Name)
		}

		time.Sleep(5 * time.Second)
	}
}
