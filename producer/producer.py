import beanstalkc
import random
import time

bs = beanstalkc.Connection(host='beanstalkd', port=11300)
while(True):
    for i in range(10):
        bs.use("tube-" + str(random.randint(0, 10)))
        bs.put("Hello World")
    time.sleep(5)