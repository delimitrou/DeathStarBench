import requests
import time

getfollow_url = 'http://localhost:31992/v1.0/invoke/dapr-social-graph/method/getfollow'
getfollower_url = 'http://localhost:31992/v1.0/invoke/dapr-social-graph/method/getfollower'
follow_url = 'http://localhost:31992/v1.0/invoke/dapr-social-graph/method/follow'
unfollow_url = 'http://localhost:31992/v1.0/invoke/dapr-social-graph/method/unfollow'


# set up user-0 to user-9
# user-0 is followed by all other users
print('------ user-0 is followed by all other users ------')
for i in range(1, 11):
    payload = {
        'send_unix_ms': int(time.time() * 1000),
        'user_id': 'user-' + str(i),
        'follow_id': 'user-0',
    }
    r = requests.post(follow_url, json=payload)
    print(r.text)

print('------ user-0 follows user-1 and user-2 ------')
for i in range(1, 3):
    payload = {
        'send_unix_ms': int(time.time() * 1000),
        'user_id': 'user-0',
        'follow_id': 'user-' + str(i),
    }
    r = requests.post(follow_url, json=payload)
    print(r.text)

print('------ user-10 follows user-0 to user-9 ------')
for i in range(1, 10):
    payload = {
        'send_unix_ms': int(time.time() * 1000),
        'user_id': 'user-10',
        'follow_id': 'user-' + str(i),
    }
    r = requests.post(follow_url, json=payload)
    print(r.text)

print('## ------------------------------ ##')
for i in range(0, 11):
    print("------ follower list of user-%d ------" %i)
    payload = {
        'send_unix_ms': int(time.time() * 1000),
        'user_ids': ['user-'+str(i)]
    }
    r = requests.get(getfollower_url, json=payload)
    print(r.text)

print('## ------------------------------ ##')
for i in range(0, 11):
    print("------ follow list of user-%d ------" %i)
    payload = {
        'send_unix_ms': int(time.time() * 1000),
        'user_ids': ['user-'+str(i)]
    }
    r = requests.get(getfollow_url, json=payload)
    print(r.text)

