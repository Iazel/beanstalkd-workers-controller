#!/bin/bash

main() {
    if [ -z "$1" ]; then
        true \
        && compile_main \
        && compile_consumer \
        && compile_producer
    fi

    "compile_$1"
}

compile_main() {
    cd compile
    local img_build='beanstalkd-workers-controller-build'
    docker build -t $img_build .

    cd ..
    if $(docker ps | grep $img_build); then
        docker start $img_build
    else
        docker run -v "$(readlink -m .):/go/src/app" -v "/tmp/gocache:/go/pkg" --name $img_build -- $img_build
    fi
    docker build -t 'beanstalkd-workers-controller:1.0' .
}

compile_consumer() {
    cd consumer
    docker build -t 'beanstalkd-consumer:1.0' .
    cd ..
}

compile_producer() {
    cd producer
    docker build -t 'beanstalkd-producer:1.0' .
    cd ..
}

main "$@"
