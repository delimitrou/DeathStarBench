import requests
import time
import json

getfollow_url = 'http://localhost:31990/v1.0/invoke/dapr-social-graph/method/getfollow'
getfollower_url = 'http://localhost:31990/v1.0/invoke/dapr-social-graph/method/getfollower'
follow_url = 'http://localhost:31990/v1.0/invoke/dapr-social-graph/method/follow'
unfollow_url = 'http://localhost:31990/v1.0/invoke/dapr-social-graph/method/unfollow'


def make_user(i):
    return 'tester-' + str(i)

# Apart from following each user himself
# tester-0 follows tester-1, tester-2 & tester-3
# tester-1 follows tester-2, tester-3 & tester-4 
# ...
# tester-4 follows tester-0, tester-1 & tester-2
# 
# tester-0 followed by tester-2, tester-3 & tester-4
# tester-1 followed by tester-3, tester-4 & tester-0 
# ...
# tester-4 followed by tester-1, tester-2 & tester-3
for i in range(0, 5):
    user = make_user(i)
    for j in range(1, 4):
        follow = (i + j) % 5
        payload = {
            'send_unix_ms': int(time.time() * 1000),
            'user_id': user,
            'follow_id': make_user(follow),
        }
        r = requests.post(follow_url, json=payload)
        print(r.text)

print('------ check follow ------')
for i in range(0, 5):
    user = make_user(i)
    payload = {
        'send_unix_ms': int(time.time() * 1000),
        'user_ids': [user],
    }
    r = requests.get(getfollow_url, json=payload)
    f = json.loads(r.text)['follow_ids'][user]
    print('%s follows %s' %(user, ', '.join(f)))

print('\n------ check follower ------')
for i in range(0, 5):
    user = make_user(i)
    payload = {
        'send_unix_ms': int(time.time() * 1000),
        'user_ids': [user],
    }
    r = requests.get(getfollower_url, json=payload)
    f = json.loads(r.text)['follower_ids'][user]
    print('%s followed by %s' %(user, ', '.join(f)))