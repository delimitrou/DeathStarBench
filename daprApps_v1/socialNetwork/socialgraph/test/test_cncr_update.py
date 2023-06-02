import requests
import time
import threading
import argparse
import json

parser = argparse.ArgumentParser()
parser.add_argument('--c', dest='concurrency', type=int, default=5)
parser.add_argument('--u', dest='updates', type=int, default=10)
args = parser.parse_args()
concurrency = args.concurrency
updates = args.updates

getfollow_url = 'http://localhost:31992/v1.0/invoke/dapr-social-graph/method/getfollow'
getfollower_url = 'http://localhost:31992/v1.0/invoke/dapr-social-graph/method/getfollower'
follow_url = 'http://localhost:31992/v1.0/invoke/dapr-social-graph/method/follow'
unfollow_url = 'http://localhost:31992/v1.0/invoke/dapr-social-graph/method/unfollow'

key = 'Chives'

def updater(id, key, updates):
    for i in range(0, updates):
        follow_id = 'sickle-%d-%d' %(id, i)
        payload = {
            'send_unix_ms': int(time.time() * 1000),
            'user_id': follow_id,
            'follow_id': key, 
        }
        r = requests.post(follow_url, json=payload)
        print(r.text)

threads = []
for i in range(0, concurrency):
    t = threading.Thread(
        target=updater, 
        kwargs={
            'id': i,
            'key': key,
            'updates': updates,
        }
    )
    threads.append(t)
    t.start()

for t in threads:
    t.join()

# assume starting from an empty store
print("------ Check follow list of Sickle ------")
for i in range(0, concurrency):
    for j in range(0, updates):
        user = 'sickle-%d-%d' %(i, j)
        payload = {
            'send_unix_ms': int(time.time() * 1000),
            'user_ids': [user]
        }
        print(user)
        r = requests.get(getfollow_url, json=payload)
        print(r.text, '\n')

print("------ Check follower list of Chives ------")
user = 'Chives'
payload = {
    'send_unix_ms': int(time.time() * 1000),
    'user_ids': [user]
}
print(user)
r = requests.get(getfollower_url, json=payload)
print(r.text, '\n')
data = json.loads(r.text)
print(len(data['follower_ids'][user]))
