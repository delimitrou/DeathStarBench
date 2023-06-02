import requests
import time

update_tl_url = 'http://localhost:31891/v1.0/publish/timeline-events/timeline'
store_url = 'http://localhost:31891/v1.0/state/timeline-store-test/'

def genPostId(user):
    return '%s*%d' %(user, int(time.time() * 1000))

def userTl(user):
    return "%s-u" %user

def homeTl(user):
    return "%s-h" %user

for p in range(0, 5):
    # assume starting from an empty store
    print("------ user-0 adds a post ------")
    payload = {
        'user_id': 'user-0',
        'post_id': genPostId('user-0'),
        'add': True,
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

for p in range(0, 4):
    # assume starting from an empty store
    u = 'user-1'
    if p % 2 == 1:
        u = 'user-2'
    print("------ %s adds a post ------" %u)
    payload = {
        'user_id': u,
        'post_id': genPostId(u),
        'add': True,
        'send_unix_ms': int(time.time() * 1000),
    }
    r = requests.post(update_tl_url, json=payload)
    print(r.text)

    time.sleep(0.1)
    print("--- %s ---" %userTl('user-0'))
    r = requests.get(store_url + userTl('user-0'))
    print(r.text)

    for i in range(0, 10):
        print("--- %s ---" %homeTl('user-%d' %i))
        r = requests.get(store_url + homeTl('user-%d' %i))
        print(r.text)

