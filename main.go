package main

import (
	"context"
	"flag"
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/klog/v2"
)

func main() {
	// Parse flags
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Parse()

	// Use in-cluster config, or use kubeconfig if provided
	config, err := rest.InClusterConfig()
	if err != nil {
		klog.Infof("Falling back to kubeconfig: %v", err)
		kubeconfig := os.Getenv("KUBECONFIG")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			klog.Fatalf("Error creating kubeconfig: %v", err)
		}
	}

	// Create Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Error creating Kubernetes client: %v", err)
	}

	// Create a Resource Lock for leader election
	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      "example-leader-election",
			Namespace: "default", // Replace with your namespace
		},
		Client: clientset.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: os.Getenv("POD_NAME"), // Pod identity is the hostname by default
		},
	}

	// Context to manage the leader election
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Leader Election configuration
	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:          lock,
		LeaseDuration: 15 * time.Second, // Lease duration before another instance can take over
		RenewDeadline: 10 * time.Second, // How often the leader renews its lease
		RetryPeriod:   2 * time.Second,  // Retry interval for leader election
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				// This function is called when the pod becomes the leader
				klog.Info("I am the leader now")
				for {
					klog.Info("Doing leader's work")
					time.Sleep(5 * time.Second)
				}
			},
			OnStoppedLeading: func() {
				// This function is called when this pod stops being the leader
				klog.Infof("Leader lost")
				cancel()
			},
		},
	})
}
