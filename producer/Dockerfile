FROM python:2.7-alpine

RUN pip install beanstalkc PyYAML
ADD producer.py /

ENTRYPOINT ["python", "/producer.py"]