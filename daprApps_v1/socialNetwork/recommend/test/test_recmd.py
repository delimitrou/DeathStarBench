import requests
import time
import json

recmd_url = 'http://localhost:31991/v1.0/invoke/dapr-recommend/method/recmd'

def make_user(i):
    return 'recmd-user-' + str(i)

for i in range(0, 20):
    print('------ follow recmd for recmd-user-%d ------' %i)
    payload = {
        'send_unix_ms': int(time.time() * 1000),
        'user_id': make_user(i),
    }
    r = requests.post(recmd_url, json=payload)
    resp = json.loads(r.text)
    # print(resp)
    assert make_user(i) not in resp['user_ids']
    print(len(resp['user_ids']))
    print(sorted(resp['user_ids']))