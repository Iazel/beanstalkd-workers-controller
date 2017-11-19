FROM alpine

ADD controller k8s/consumer.yml /app/
WORKDIR /app

ENTRYPOINT ["./controller"]
