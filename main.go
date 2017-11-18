package main

import (
	"fmt"
	"log"
	"time"

	"github.com/iwanbk/gobeanstalk"
	"gopkg.in/yaml.v2"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
)

func int32Ptr(i int32) *int32 { return &i }

type stats struct {
	Ready    int `yaml:"current-jobs-ready"`
	Watching int `yaml:"current-watching"`
}

func main() {
	k8s()
}

func _main() {
	bs, err := gobeanstalk.Dial("localhost:11300")
	if err != nil {
		log.Fatal(err)
	}

	for {
		tubesStats(bs, func(stats *stats) {
			fmt.Printf("Jobs ready: %d\n", stats.Ready)
			fmt.Printf("Watching: %d\n", stats.Watching)
		})
		time.Sleep(1 * time.Second)
	}
}

func tubesStats(bs *gobeanstalk.Conn, do func(*stats)) {
	tubesYaml, err := bs.ListTubes()
	if err != nil {
		log.Fatal(err)
	}

	var tubes []string
	err = yaml.UnmarshalStrict(tubesYaml, &tubes)
	if err != nil {
		log.Fatal(err)
	}

	for _, t := range tubes {
		statsYaml, err := bs.StatsTube(t)
		if err != nil {
			log.Fatal(err)
		}

		var stats stats
		err = yaml.Unmarshal(statsYaml, &stats)
		if err != nil {
			log.Fatal(err)
		}
		do(&stats)
	}
}

func k8s() {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	for {
		rsClient := clientset.
			AppsV1beta2().
			ReplicaSets(apiv1.NamespaceDefault)

		retry.RetryOnConflict(retry.DefaultRetry, func() error {
			rs, err := rsClient.Get("test", metav1.GetOptions{})
			if err != nil {
				panic(fmt.Errorf("Failed to get latest version of Deployment: %v", err))
			}

			rs.Spec.Replicas = int32Ptr(*rs.Spec.Replicas + 1)
			_, err = rsClient.Update(rs)
			return err
		})

		pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

		// Examples for error handling:
		// - Use helper functions like e.g. errors.IsNotFound()
		// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
		_, err = clientset.CoreV1().Pods("default").Get("example-xxxxx", metav1.GetOptions{})
		if errors.IsNotFound(err) {
			fmt.Printf("Pod not found\n")
		} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
			fmt.Printf("Error getting pod %v\n", statusError.ErrStatus.Message)
		} else if err != nil {
			panic(err.Error())
		} else {
			fmt.Printf("Found pod\n")
		}

		time.Sleep(10 * time.Second)
	}
}
