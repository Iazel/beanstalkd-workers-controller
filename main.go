package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"

	v1beta1 "k8s.io/api/extensions/v1beta1"
	extv1b1 "k8s.io/client-go/kubernetes/typed/extensions/v1beta1"

	"github.com/iwanbk/gobeanstalk"
	"gopkg.in/yaml.v2"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
)

func checkFatalError(msg string, err error) {
	if err != nil {
		panic(fmt.Errorf(msg+": %v", err))
	}
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
	checkFatalError("Can't initialize Beanstalkd", err)
	return bs
}

func tubesStats(bs *gobeanstalk.Conn, do func(string, *stats)) {
	tubesYaml, err := bs.ListTubes()
	checkFatalError("Can't retrieve tubes", err)

	var tubes []string
	err = yaml.UnmarshalStrict(tubesYaml, &tubes)
	checkFatalError("Wrong tube response", err)

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
			rs, finalize, err := getReplicaSet(rsClient, tube)
			if err != nil {
				log.Printf("Tube `%s` - failed to get ReplicaSet: %v\n", tube, err)
				return err
			}

			rs.Spec.Replicas = calcReplicas(stats.Ready)
			log.Printf(
				"\nTube %s\nJobs ready: %d\nSetting replicas to: %d\n\n",
				tube, stats.Ready, *rs.Spec.Replicas)
			_, err = finalize(rs)
			return err
		})
	}
}

func getReplicaSet(rsClient extv1b1.ReplicaSetInterface, tube string) (*v1beta1.ReplicaSet, func(*v1beta1.ReplicaSet) (*v1beta1.ReplicaSet, error), error) {
	replicaName := "consumer-" + tube
	rs, err := rsClient.Get(replicaName, metav1.GetOptions{})
	if err == nil {
		return rs, rsClient.Update, nil
	}

	rs, err = setupReplicaSet(rsClient, tube, replicaName)
	if err != nil {
		return nil, nil, err
	}
	return rs, rsClient.Create, nil
}

func setupReplicaSet(rsClient extv1b1.ReplicaSetInterface, tube string, name string) (*v1beta1.ReplicaSet, error) {
	yaml, err := ioutil.ReadFile("./consumer.yml")
	if err != nil {
		return nil, err
	}

	schema, _, err := scheme.Codecs.UniversalDeserializer().Decode(yaml, nil, nil)
	rsSchema := castReplicaSetSchema(schema)
	rsSchema.ObjectMeta.Name = name
	rsSchema.Spec.Template.Spec.Containers[0].Env[0].Value = tube
	return rsSchema, nil
}

func castReplicaSetSchema(schema interface{}) *v1beta1.ReplicaSet {
	rsSchema, ok := schema.(*v1beta1.ReplicaSet)
	if ok {
		return rsSchema
	}
	panic("Schema is not a ReplicaSet!")
}

func calcReplicas(ready int32) *int32 {
	// n default to 0
	var n int32
	if ready == 0 {
		return &n
	}

	n = lowerBound(1, ready/10)
	return &n
}

func lowerBound(limit int32, n int32) int32 {
	if n <= limit {
		return limit
	}
	return n
}

func initReplicaSetsClient() extv1b1.ReplicaSetInterface {
	config, err := rest.InClusterConfig()
	checkFatalError("Can't access k8s API", err)

	clientset, err := kubernetes.NewForConfig(config)
	checkFatalError("Can't initialize k8s client", err)

	return clientset.
		ExtensionsV1beta1().
		ReplicaSets(v1.NamespaceDefault)
}
