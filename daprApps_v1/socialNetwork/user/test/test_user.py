import requests
import time
import json

reg_url = 'http://localhost:31991/v1.0/invoke/dapr-user/method/register'
getfollow_url = 'http://localhost:31992/v1.0/invoke/dapr-social-graph/method/getfollow'
getfollower_url = 'http://localhost:31992/v1.0/invoke/dapr-social-graph/method/getfollower'

def make_user(i):
    return 'reg-user-' + str(i)

for i in range(0, 20):
    print('------ register reg-user-%d ------' %i)
    payload = {
        'send_unix_ms': int(time.time() * 1000),
        'user_id': make_user(i),
    }
    r = requests.post(reg_url, json=payload)
    print(r.text)

for i in range(0, 20):
    print('-------- %s --------' %make_user(i))
    print('-- follows --')
    payload = {
        'send_unix_ms': int(time.time() * 1000),
        'user_ids': [make_user(i)]
    }
    r = requests.get(getfollow_url, json=payload)
    print(r.text)

    print('-- followers --')
    r = requests.get(getfollower_url, json=payload)
    print(r.text)
