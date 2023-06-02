import requests
import time
import json

reg_url = 'http://localhost:31991/v1.0/invoke/dapr-user/method/register'

def make_user(i):
    return 'tester-' + str(i)

for i in range(0, 5):
    print('------ register reg-user-%d ------' %i)
    payload = {
        'send_unix_ms': int(time.time() * 1000),
        'user_id': make_user(i),
    }
    r = requests.post(reg_url, json=payload)
    print(r.text)
