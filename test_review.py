import urllib.request
import json

data = json.dumps({"grade": 1}).encode('utf-8')
req = urllib.request.Request(
    'http://localhost:8081/api/v1/questions/9e956f6c-2178-4dd7-91c2-7cc8756cdabe/review',
    data=data,
    headers={'Content-Type': 'application/json'}
)
print(urllib.request.urlopen(req).read().decode('utf-8'))
