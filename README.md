# Kubernetes beanstalkd workers

## Disclaimer
This is just a demo and nothing more than a proof of concept.

## Goal
Dynamically provision and scale consumers based on beanstalkd tubes and number of ready jobs.  

The controller has only two expectations for consumers:
* They are managed through ReplicaSet
* They read from a tube specified in the env variable `QUEUE`

All other details can be easily customizable through `k8s/consumer.yml`

## Requirements
* minikube
* kubectl (with kube-dns activated)

## Setup
```sh
cd path/to/this/project
# mount project dir in minikube
minikube mount "$(readlink -m .):/home/app"
```
Leave the terminal open and in another one:
```sh
minikube ssh
```
Once you are inside minikube VM:
```sh
cd /home/app
bash compile.sh
```
Wait until completition and then exit from minikube.  
At this point you may want to close `minikube mount` terminal too

## Run
```sh
cd k8s
# It's important to run beanstalkd enabled
kubectl -f beanstalkd.yml
# Wait until beanstalkd is up and running, should be a matter of seconds
kubectl -f workers-controller.yml
kubectl -f producer.yml
```
As you can see we didn't run any consumer because workers-controller will take care of spawning them automatically as needed.

After a couple of seconds, you should see consumers up and running by either using kubernetes dashboard or:
```sh
kubectl get pods | grep consumer
```
