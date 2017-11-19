import os
import time
import random
import beanstalkc

tube = os.environ['QUEUE']
bs = beanstalkc.Connection(host='beanstalkd', port=11300)
bs.use(tube)

while(True):
    bs.reserve().delete()
    print "Remove job from " + tube
    time.sleep(random.randint(1, 5))
