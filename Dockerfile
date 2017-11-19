FROM alpine

ADD bin/controller k8s/consumer.yml /app/
WORKDIR /app

ENTRYPOINT ["./controller"]
