import requests
import time
import json

update_tl_url = 'http://localhost:31891/v1.0/publish/timeline-events/timeline'
store_url = 'http://localhost:31891/v1.0/state/timeline-store-test/'

def genPostId(user):
    return '%s*%d' %(user, int(time.time() * 1000))

def userTl(user):
    return "%s-u" %user

def homeTl(user):
    return "%s-h" %user

# get posts of user-0
print("--- %s ---" %userTl('user-0'))
r = requests.get(store_url + userTl('user-0'))
posts = json.loads(r.text)
print(posts)
dels = int(len(posts)/2)
for pos in range(0, dels):
    p = posts[pos]
    # assume starting from an empty store
    print("------ user-0 del post %s ------" %p)
    payload = {
        'user_id': 'user-0',
        'post_id': p,
        'add': False,
        'send_unix_ms': int(time.time() * 1000),
    }
    r = requests.post(update_tl_url, json=payload)
    print(r.text)

    time.sleep(0.1)
    print("--- %s ---" %userTl('user-0'))
    r = requests.get(store_url + userTl('user-0'))
    print(r.text)

    for i in range(1, 10):
        print("--- %s ---" %homeTl('user-%d' %i))
        r = requests.get(store_url + homeTl('user-%d' %i))
        print(r.text)

