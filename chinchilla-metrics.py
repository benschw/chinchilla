#!/usr/bin/env python

import urllib, json, time, socket
url = "http://localhost:8081/debug/vars"

response = urllib.urlopen(url)
data = json.loads(response.read())


host = socket.gethostname()
ts = time.time()

if data["metrics"]:
    for k in data["metrics"]:
        for metricKey, val in data["metrics"][k].iteritems():
            print("%s.%s\t%d\t %d" % (host, metricKey, val, ts))


