import requests
import time
import threading
import argparse
import json
from typing import Optional, Union, Callable, List

parser = argparse.ArgumentParser()
parser.add_argument('--c', dest='concurrency', type=int, default=5)
parser.add_argument('--u', dest='updates', type=int, default=10)
args = parser.parse_args()
concurrency = args.concurrency
updates = args.updates

save_url = 'http://localhost:31991/v1.0/invoke/dapr-post/method/save'
del_url = 'http://localhost:31991/v1.0/invoke/dapr-post/method/del'
meta_url = 'http://localhost:31991/v1.0/invoke/dapr-post/method/meta'
comment_url = 'http://localhost:31991/v1.0/invoke/dapr-post/method/comment'
upvote_url = 'http://localhost:31991/v1.0/invoke/dapr-post/method/upvote'
read_url = 'http://localhost:31991/v1.0/invoke/dapr-post/method/read'
# state store url
state_url = 'http://localhost:31991/v1.0/state/post-store-test/'
key = 'Chives'

# helper functions
def make_save_post(post_id: str, user_id: str, text: str, images: List[str]):
    cont = {
        'user_id': user_id,
        'text': text,
        'images': images,
    }
    return {
        'send_unix_ms': int(time.time() * 1000),
        'post_id': post_id,
        'content': cont,
    }

def make_meta(post_id: str, sent: Optional[str]=None, objects: Optional[dict]=None):
    payload = {
        'post_id': post_id,
    }
    if sent != None:
        payload['sentiment'] = sent
    if objects != None:
        payload['objects'] = objects
    payload['send_unix_ms'] = int(time.time() * 1000)
    return payload

def make_read(post_ids: List[str]):
    return {
        'post_ids': post_ids,
        'send_unix_ms': int(time.time() * 1000),
    }

def make_comment(post_id: str, user_id: str, comm_id: str, reply_to: str, text: str):
    comm = {
        'comment_id': comm_id,
        'user_id': user_id,
        'reply_to': reply_to,
        'text': text,
    }
    return {
        'post_id': post_id,
        'comm': comm,
        'send_unix_ms': int(time.time() * 1000),
    }

def make_upvote(post_id: str, user_id: str):
    return {
        'post_id': post_id,
        'user_id': user_id,
        'send_unix_ms': int(time.time() * 1000),
    }

def show_posts(postsjson: str):
    postsdata = json.loads(postsjson)
    posts = postsdata['posts']
    for post_id in posts:
        post = posts[post_id]
        print('-- post_id:', post_id)
        print('------ content:', post['content'])
        print('------ meta:', post['meta'])
        print('------ comments:', post['comments'])
        print('------ upvotes:', post['upvotes'])
    if not post['comments']['comments'] is None:
        print('#comments = ', len(post['comments']['comments']))

def update_comm(updates: int, post_id: str, user_id: str):
    ctr = 0
    for i in range(0, updates):
        comm_id = 'post-id-' + user_id + '-' + str(ctr)
        ctr += 1
        text = 'Get out of acamdemia! X' + str(ctr)
        comm_req = make_comment(
            post_id=post_id, 
            user_id=user_id,
            comm_id=comm_id,
            reply_to='',
            text=text)
        r = requests.post(comment_url, json=comm_req)

post_id = 'cncr-comm-post-1'
print("------ save post contents ------")
save_req = make_save_post(
    post_id=post_id, 
    user_id='Integrity', 
    text='Fakers get out of academia!', 
    images=['FGW.jpg'],
)
r = requests.post(save_url, json=save_req)
print(r.text)

print("------ read post ------")
read_req = make_read(post_ids=[post_id])
r = requests.get(read_url, json=read_req)
show_posts(r.text)

threads = []
for i in range(0, concurrency):
    t = threading.Thread(
        target= update_comm,
        kwargs={
            'updates': updates,
            'post_id': post_id,
            'user_id': 'ZhenShiYin-' + str(i),
        }
    )
    threads.append(t)
    t.start()

for t in threads:
    t.join()

print("------ read post (after comments updated) ------")
read_req = make_read(post_ids=[post_id])
r = requests.get(read_url, json=read_req)
show_posts(r.text)


