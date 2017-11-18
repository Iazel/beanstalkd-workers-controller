package main

import (
	"fmt"
	"log"
	"time"

	"github.com/iwanbk/gobeanstalk"
	"gopkg.in/yaml.v2"
	v1beta2 "k8s.io/api/apps/v1beta2"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appv1b2 "k8s.io/client-go/kubernetes/typed/apps/v1beta2"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
)

func panicError(msg string, err error) {
	panic(fmt.Errorf(msg+": %v", err))
}

type stats struct {
	Ready    int32 `yaml:"current-jobs-ready"`
	Watching int32 `yaml:"current-watching"`
}

func main() {
	bs := initBeanstalkd()
	defer bs.Quit()

	spawner := initSpawner()

	for {
		tubesStats(bs, spawner)
		time.Sleep(5 * time.Second)
	}
}

func initBeanstalkd() *gobeanstalk.Conn {
	bs, err := gobeanstalk.Dial("beanstalkd:11300")
	if err != nil {
		panicError("Can't initialize Beanstalkd", err)
	}
	return bs
}

func tubesStats(bs *gobeanstalk.Conn, do func(string, *stats)) {
	tubesYaml, err := bs.ListTubes()
	if err != nil {
		panicError("Can't retrieve tubes", err)
	}

	var tubes []string
	err = yaml.UnmarshalStrict(tubesYaml, &tubes)
	if err != nil {
		panicError("Wrong tube response", err)
	}

	for _, t := range tubes {
		statsYaml, err := bs.StatsTube(t)
		if err != nil {
			log.Printf("Tube `%s` - can't fetch stats: %v\n", t, err)
			continue
		}

		var stats stats
		err = yaml.Unmarshal(statsYaml, &stats)
		if err != nil {
			log.Printf("Tube `%s` - can't unmarshal: %v\n", t, err)
			continue
		}
		do(t, &stats)
	}
}

func initSpawner() func(string, *stats) {
	rsClient := initReplicaSetsClient()

	return func(tube string, stats *stats) {
		retry.RetryOnConflict(retry.DefaultRetry, func() error {
			rs, err := getReplicaSet(rsClient, tube)
			if err != nil {
				log.Printf("Tube `%s` - failed to get ReplicaSet: %v\n", tube, err)
				return err
			}

			rs.Spec.Replicas = calcReplicas(stats.Ready)
			_, err = rsClient.Update(rs)
			return err
		})
	}
}

func getReplicaSet(rsClient appv1b2.ReplicaSetInterface, tube string) (*v1beta2.ReplicaSet, error) {
	replicaName := "producer-" + tube
	rs, err := rsClient.Get(replicaName, metav1.GetOptions{})
	if err == nil {
		return rs, nil
	}
	return nil, err
}

func int32Ptr(i int32) *int32 {
	return &i
}

func calcReplicas(ready int32) *int32 {
	// n default to 0
	var n int32
	if ready == 0 {
		return &n
	}

	n = ready / 10
	return &n
}

func initReplicaSetsClient() appv1b2.ReplicaSetInterface {
	config, err := rest.InClusterConfig()
	if err != nil {
		panicError("Can't access k8s API", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panicError("Can't initialize k8s client", err)
	}

	return clientset.
		AppsV1beta2().
		ReplicaSets(apiv1.NamespaceDefault)
}
