package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	// 读取参数
	business := flag.String("business", "", "business label value (e.g. devops)")
	namespace := flag.String("namespace", "", "namespace to query (default: all)")
	flag.Parse()

	if *business == "" {
		log.Fatal("必须指定 --business 参数，例如: --business=devops")
	}

	// kubeconfig 路径
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	// 创建配置
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("构建 kubeconfig 失败: %v", err)
	}

	// 创建客户端
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("创建 Kubernetes client 失败: %v", err)
	}

	ctx := context.Background()
	labelSelector := fmt.Sprintf("business=%s", *business)

	// 查询 Deployments
	deployments, err := clientset.AppsV1().Deployments(*namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		log.Fatalf("获取 Deployments 失败: %v", err)
	}

	// 查询 StatefulSets
	statefulsets, err := clientset.AppsV1().StatefulSets(*namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		log.Fatalf("获取 StatefulSets 失败: %v", err)
	}

	// 打印结果
	fmt.Println("=== Deployments ===")
	for _, d := range deployments.Items {
		fmt.Printf("%s/%s\n", d.Namespace, d.Name)
	}

	fmt.Println("=== StatefulSets ===")
	for _, s := range statefulsets.Items {
		fmt.Printf("%s/%s\n", s.Namespace, s.Name)
	}
}
