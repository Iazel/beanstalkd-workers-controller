import os
import time
import random
import beanstalkc

bs = beanstalkc.Connection(host='beanstalkd', port=11300)
bs.use(os.environ['QUEUE'])

while(True):
    bs.reserve().delete()
    time.sleep(random.randint(1, 5))