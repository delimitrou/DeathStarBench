import requests
import time
import threading
import argparse
import json

parser = argparse.ArgumentParser()
parser.add_argument('--p', dest='posts', type=int, default=5)
args = parser.parse_args()
posts = args.posts

update_tl_url = 'http://localhost:31891/v1.0/publish/timeline-events/timeline'
store_url = 'http://localhost:31891/v1.0/state/timeline-store-test/'
update_users = 10

def genPostId(user):
    return '%s*%d' %(user, int(time.time() * 1000))

def userTl(user):
    return "%s-u" %user

def homeTl(user):
    return "%s-h" %user

def updater(user, posts):
    for i in range(0, posts):
        payload = {
            'user_id': user,
            'post_id': genPostId(user),
            'add': True,
            'send_unix_ms': int(time.time() * 1000),
        }
        r = requests.post(update_tl_url, json=payload)
        # print(r.text)

threads = []
for i in range(0, 10):
    t = threading.Thread(
        target=updater, 
        kwargs={
            'user': 'user-' + str(i),
            'posts': posts,
        }
    )
    threads.append(t)
    t.start()

for t in threads:
    t.join()

time.sleep(0.1)
print("--- %s ---" %homeTl('user-10'))
r = requests.get(store_url + homeTl('user-10'))
posts = json.loads(r.text)
print(posts)
print(len(posts))

unix_ms = [int(p.split('*')[-1]) for p in posts]
for i in range(0, len(unix_ms) - 1):
    assert unix_ms[i] <= unix_ms[i+1]

