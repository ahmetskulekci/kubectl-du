package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	namespace string
	nodeName  string
)

var rootCmd = &cobra.Command{
	Use:   "disk",
	Short: "Display disk usage of pods on a Kubernetes node",
	Run: func(cmd *cobra.Command, args []string) {
		kubeconfig := os.Getenv("KUBECONFIG")
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			config, err = rest.InClusterConfig()
			if err != nil {
				fmt.Printf("Error creating config: %v\n", err)
				os.Exit(1)
			}
		}

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			fmt.Printf("Error creating Kubernetes client: %v\n", err)
			os.Exit(1)
		}

		if nodeName == "" {
			listNodesDiskUsage(clientset)
		} else {
			getPodDiskUsage(clientset, nodeName, namespace)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "ns", "", "Specify the namespace to filter pods")
	rootCmd.PersistentFlags().StringVarP(&nodeName, "node", "", "", "Specify the node to check disk usage")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func getPodDiskUsage(clientset *kubernetes.Clientset, nodeName string, namespace string) {
	fieldSelector := fmt.Sprintf("spec.nodeName=%s", nodeName)
	listOptions := metav1.ListOptions{
		FieldSelector: fieldSelector,
	}

	if namespace != "" {
		pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), listOptions)
		if err != nil {
			fmt.Printf("Error getting pods: %v\n", err)
			return
		}
		displayPods(pods)
	} else {
		pods, err := clientset.CoreV1().Pods("").List(context.TODO(), listOptions)
		if err != nil {
			fmt.Printf("Error getting pods: %v\n", err)
			return
		}
		displayPods(pods)
	}
}

func listNodesDiskUsage(clientset *kubernetes.Clientset) {
	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error getting nodes: %v\n", err)
		return
	}

	for _, node := range nodes.Items {
		fmt.Printf("Node: %s\n", node.Name)
		getPodDiskUsage(clientset, node.Name, namespace)
	}
}

func displayPods(pods *v1.PodList) { // Burada v1.PodList kullanıldı
	for _, pod := range pods.Items {
		fmt.Printf("Pod: %s, Namespace: %s\n", pod.Name, pod.Namespace)
		fmt.Printf("Volume Count: %d\n", len(pod.Spec.Volumes))
		fmt.Println(strings.Repeat("-", 40))
	}
}
