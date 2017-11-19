FROM python:2.7-alpine

RUN pip install beanstalkc PyYAML
ADD consumer.py /

ENTRYPOINT ["python", "/consumer.py"]
