package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {

	chief := []string{
		"chief-api",
		"chief-common-api",
		"chief-common-support",
		"chief-common-support-admin-server",
		"chief-custom",
		"chief-ekyc",
		"chief-fund",
		"chief-fundaccount-va",
		"chief-fund-admin",
		"chief-gateway",
		"chief-ipo-server",
		"chief-management",
		"chief-message-push",
		"chief-message-server",
		"chief-news",
		"chief-operations",
		"chief-operations-admin",
		"chief-sentinel-dashboard",
		"chief-server",
		"chief-sso-admin-server",
		"chief-toptrader-ae",
		"chief-toptrader-speed",
		"chief-trade",
		"chief-trader-crypto",
		"chief-trader-fundserver",
		"chief-trader-x",
		"chief-traderx-admin",
		"customer-info-admin-server",
		"customer-info-server",
		"ekyc-admin-server",
		"hkfuturesenlarge51",
		"hk-sso-server",
		"pushmsgforgold",
		"sso-server",
		"szca",
		"trade-crypto-admin",
		"chief-data-report",
		"chief-common-job",
		"chief-stresstest-locust",
		"chief-trader-data-sync",
		"datax-sync",
	}

	quote := []string{
		"chief-external",
		"chief-quote",
		"chief-quote-data",
		"chief-quote-engine",
		"external-usoption",
		"index-us-external",
		"quote-api",
		"quote-data-aggregate-server",
		"quote-engine-aq",
		"quote-engine-hk-dispatcher",
		"quote-engine-hk-mdf",
		"quote-engine-hk-partition",
		"quote-engine-us",
		"quote-engine-us-external-fintech",
		"quote-engine-us-external-ice",
		"quote-engine-us-external-mdf",
		"quote-gateway-server",
		"quote-hk",
		"quote-index-futures",
		"quote-us",
		"quote-usoption",
		"quote-usoption-zfb",
	}

	other := ""

	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("build kubeconfig failed: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("create client failed: %v", err)
	}

	ctx := context.Background()

	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Fatalf("list namespaces failed: %v", err)
	}

	for _, ns := range namespaces.Items {
		nsName := ns.Name

		// Deployment
		deployments, _ := clientset.AppsV1().Deployments(nsName).List(ctx, metav1.ListOptions{})
		for _, d := range deployments.Items {
			updateBusiness(ctx, clientset, nsName, d.Name, d.Labels, chief, quote, other)
		}

		// StatefulSet
		statefulsets, _ := clientset.AppsV1().StatefulSets(nsName).List(ctx, metav1.ListOptions{})
		for _, s := range statefulsets.Items {
			updateBusiness(ctx, clientset, nsName, s.Name, s.Labels, chief, quote, other)
		}
	}
}

// updateBusiness 更新 Deployment/STS、对应 Pod 和 Service 的 business 标签
func updateBusiness(ctx context.Context, clientset *kubernetes.Clientset, ns, name string, labelsMap map[string]string, chief, quote []string, other string) {
	if labelsMap == nil {
		labelsMap = map[string]string{}
	}

	old := labelsMap["business"]
	newBusiness := other
	if contains(chief, name) {
		newBusiness = "chief"
	} else if contains(quote, name) {
		newBusiness = "quote"
	}

	if old == newBusiness {
		return
	}

	// 更新 Deployment/STS
	labelsMap["business"] = newBusiness

	// 先尝试更新 Deployment
	deploy, err := clientset.AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		deploy.Labels = labelsMap
		clientset.AppsV1().Deployments(ns).Update(ctx, deploy, metav1.UpdateOptions{})
		fmt.Printf("Updated Deployment %s/%s business=%s\n", ns, name, newBusiness)
	}

	// 再尝试更新 StatefulSet
	sts, err := clientset.AppsV1().StatefulSets(ns).Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		sts.Labels = labelsMap
		clientset.AppsV1().StatefulSets(ns).Update(ctx, sts, metav1.UpdateOptions{})
		fmt.Printf("Updated StatefulSet %s/%s business=%s\n", ns, name, newBusiness)
	}

	// 更新 Pod
	selector := labels.SelectorFromSet(labelsMap)
	pods, _ := clientset.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{LabelSelector: selector.String()})
	for _, pod := range pods.Items {
		if pod.Labels == nil {
			pod.Labels = map[string]string{}
		}
		pod.Labels["business"] = newBusiness
		clientset.CoreV1().Pods(ns).Update(ctx, &pod, metav1.UpdateOptions{})
		fmt.Printf("Updated Pod %s/%s business=%s\n", ns, pod.Name, newBusiness)
	}

	// 更新 Service
	services, _ := clientset.CoreV1().Services(ns).List(ctx, metav1.ListOptions{})
	for _, svc := range services.Items {
		if svc.Labels == nil {
			svc.Labels = map[string]string{}
		}
		// 判断 service selector 是否匹配
		if selector.Matches(labels.Set(svc.Spec.Selector)) {
			svc.Labels["business"] = newBusiness
			clientset.CoreV1().Services(ns).Update(ctx, &svc, metav1.UpdateOptions{})
			fmt.Printf("Updated Service %s/%s business=%s\n", ns, svc.Name, newBusiness)
		}
	}
}

func contains(list []string, name string) bool {
	for _, v := range list {
		if v == name {
			return true
		}
	}
	return false
}
