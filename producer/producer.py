import beanstalkc
import random
import time

bs = beanstalkc.Connection(host='beanstalkd', port=11300)
while(True):
    for i in range(3):
        tube = "tube-" + str(random.randint(0, 5))
        bs.use(tube)
        bs.put("Hello World")
        print "Pushed a job to " + tube
    time.sleep(5)
