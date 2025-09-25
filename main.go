package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	// 构建 kubeconfig 路径
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	// 创建配置
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %v", err)
	}

	// 创建客户端
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}

	ctx := context.Background()

	// 获取所有 namespace
	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error listing namespaces: %v", err)
	}

	fmt.Println("查找包含 'business' 的服务:")
	fmt.Println("================================")

	for _, ns := range namespaces.Items {
		nsName := ns.Name

		// 查找 Deployments
		deployments, err := clientset.AppsV1().Deployments(nsName).List(ctx, metav1.ListOptions{})
		if err != nil {
			log.Printf("Error listing deployments in namespace %s: %v", nsName, err)
			continue
		}

		for _, deploy := range deployments.Items {
			if strings.Contains(deploy.Name, "business") {
				fmt.Printf("Deployment: %s/%s\n", nsName, deploy.Name)
			}
		}

		// 查找 StatefulSets
		statefulsets, err := clientset.AppsV1().StatefulSets(nsName).List(ctx, metav1.ListOptions{})
		if err != nil {
			log.Printf("Error listing statefulsets in namespace %s: %v", nsName, err)
			continue
		}

		for _, sts := range statefulsets.Items {
			if strings.Contains(sts.Name, "business") {
				fmt.Printf("StatefulSet: %s/%s\n", nsName, sts.Name)
			}
		}
	}
}
