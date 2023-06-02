import requests
import time
import json
read_tl_url = 'http://localhost:31991/v1.0/invoke/dapr-timeline-read/method/read'
store_url = 'http://localhost:31991/v1.0/state/post-store-test/'

def genPostId(user):
    return '%s*%d' %(user, int(time.time() * 1000))

def userTl(user):
    return "%s-u" %user

def homeTl(user):
    return "%s-h" %user

def postId(user, unix_ms):
    return '%s*%d' %(user, unix_ms)

def postIdTime(post_id: str):
    return int(post_id.split('*')[-1])

print('#----------------- Test user timeline -----------------#')
# ---- store to db directly and then read ---- #
user = "test-read"
unix_millis = []
posts = []
for i in range(0, 10):
    unix_ms = int(time.time() * 1000)
    post_id = postId(user, unix_ms)
    posts.append(post_id)
    time.sleep(0.5)
item = {
    'key': userTl(user),
    'value': posts,
}
r = requests.post(store_url, json=[item])
print(r.text)

# get timeline and check
print('--------- usertimeline of %s ----------' %user)
r = requests.get(store_url + userTl(user))
print(r.text)

for i, post_id in enumerate(posts):
    ts = postIdTime(post_id)
    ts += 10
    # assume starting from an empty store
    print("------ earliest ts=%d, posts=3 ------" %ts)
    payload = {
        'user_id': user,
        'user_tl': True,
        'latest_unix_milli': ts,
        'posts': 3,
        'send_unix_ms': int(time.time() * 1000),
    }
    r = requests.get(read_tl_url, json=payload)
    print(r.text)
    d = json.loads(r.text)
    print(
        len(d['post_ids']), 
        min(max(len(posts)-i-1, 0), 3),
    )
    for p in d['post_ids']:
        assert ts <= postIdTime(p)

print('#----------------- Test home timeline -----------------#')
# ---- store to db directly and then read ---- #
unix_millis = []
posts = []
for i in range(0, 10):
    unix_ms = int(time.time() * 1000)
    post_id = postId(user, unix_ms)
    posts.append(post_id)
    time.sleep(0.5)
item = {
    'key': homeTl(user),
    'value': posts,
}
r = requests.post(store_url, json=[item])
print(r.text)

# get timeline and check
print('--------- hometimeline of %s ----------' %user)
r = requests.get(store_url + homeTl(user))
print(r.text)

for i, post_id in enumerate(posts):
    ts = postIdTime(post_id)
    ts += 10
    # assume starting from an empty store
    print("------ earliest ts=%d, posts=3 ------" %ts)
    payload = {
        'user_id': user,
        'user_tl': False,
        'earl_unix_milli': ts,
        'posts': 3,
        'send_unix_ms': int(time.time() * 1000),
    }
    r = requests.get(read_tl_url, json=payload)
    print(r.text)
    d = json.loads(r.text)
    print(
        len(d['post_ids']), 
        min(max(len(posts)-i-1, 0), 3),
    )
    for p in d['post_ids']:
        assert ts <= postIdTime(p)


