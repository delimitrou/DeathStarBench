import os
import sys
import requests
import time
import json
import base64
from pathlib import Path
import random
from PIL import Image
import io

save_url = 'http://localhost:31989/v1.0/invoke/dapr-socialnet-frontend/method/save'
del_url = 'http://localhost:31989/v1.0/invoke/dapr-socialnet-frontend/method/del'
comment_url = 'http://localhost:31989/v1.0/invoke/dapr-socialnet-frontend/method/comment'
upvote_url = 'http://localhost:31989/v1.0/invoke/dapr-socialnet-frontend/method/upvote'
image_url = 'http://localhost:31989/v1.0/invoke/dapr-socialnet-frontend/method/image'
timeline_url = 'http://localhost:31989/v1.0/invoke/dapr-socialnet-frontend/method/timeline'

from pathlib import Path
img_path = Path(__file__).parent.resolve() / '..' / 'data'
sys.path.append(str(img_path))

# read images and encode as base64
images = {}
b64_images = {}
for img in os.listdir(str(img_path)):
    with open(str(img_path / img), 'rb') as f:
        images[img] = f.read()
        b64_images[img] = base64.b64encode(images[img])

# Apart from following each user himself
# tester-0 follows tester-1, tester-2 & tester-3
# tester-1 follows tester-2, tester-3 & tester-4 
# tester-2 follows tester-3, tester-4 & tester-0 
# ...
# tester-4 follows tester-0, tester-1 & tester-2
# 
# tester-0 followed by tester-2, tester-3 & tester-4
# tester-1 followed by tester-3, tester-4 & tester-0 
# ...
# tester-4 followed by tester-1, tester-2 & tester-3
def make_user(i: int):
    return 'tester-' + str(i)

def save_post(user: str, num_img: int):
    post_images = []
    img_list = list(images.keys())
    random.shuffle(img_list)
    sel_img = img_list[:num_img]
    for img in sel_img:
        post_images.append(b64_images[img])
    unix_ms = int(time.time() * 1000)
    text = '%s shouts out at %d: Fakers get out of academia!' %(user, unix_ms)
    payload = {
        'user_id': user,
        'text': text,
        'images': post_images,
        'send_unix_ms': int(time.time() * 1000),
    }
    r = requests.post(save_url, json=payload)
    return json.loads(r.text)['post_id']

def del_post(user: str, post: str):
    payload = {
        'user_id': user,
        'post_id': post,
        'send_unix_ms': int(time.time() * 1000),
    }
    r = requests.post(del_url, json=payload)
    return json.loads(r.text)['post_id']

def read_tl(user: str, earl_unix_ms: int, num_posts: int, user_tl: bool):
    payload = {
        'user_id': user,
        'user_tl': user_tl,
        'earl_unix_milli': earl_unix_ms,
        'posts': num_posts,
        'send_unix_ms': int(time.time() * 1000),
    }
    r = requests.get(timeline_url, json=payload)
    # print(r.text)
    return json.loads(r.text)

def print_tl(tl: dict):
    print('-- num_posts: %d' %len(tl['posts']))
    for post_id in tl['posts']:
        print_post(tl['posts'][post_id])
    print('########################')

def print_tl_post(tl: dict, post: str, user: str):
    if post not in tl['posts']:
        print('Error: post %s not in timeline of user %s' %(
            post, user,
        ))
    else:
        print_post(tl['posts'][post])
    print('########################')

def tl_last_post(tl: dict):
    all_posts = list(tl['posts'].keys())
    all_posts = sorted(all_posts)
    return all_posts[-1]

def print_post(post: dict):
    print('-- post_id: %s' %(post['post_id']))
    print('---- contents:')
    print('------ user_id: %s' %post['content']['user_id'])
    print('------ text: %s' %post['content']['text'])
    if post['content']['images'] != None:
        print('------ images: %s' %(','.join(post['content']['images'])))
    print('---- meta:')
    if post['meta']['sentiment'] != None:
        print('------ sentiment: %s' %(post['meta']['sentiment']))
    if post['meta']['objects'] != None:
        print('------ objects:')
        for img in post['meta']['objects']:
            print('-------- %s: %s' %(img, post['meta']['objects'][img]))
    if post['comments']['comments'] != None:
        print('---- comments:')
        for com in post['comments']['comments']:
            print('------ comment_id: %s, user_id: %s, reply_to: %s, text: %s' %(
                com['comment_id'], com['user_id'], com['reply_to'], com['text'])
            )
    if post['upvotes'] != None:
        print('---- upvotes: %s' %','.join(post['upvotes']))

def upvote_tl_post(tl: dict, user: str, post: str):
    print('-- num_posts: %d' %len(tl['posts']))
    if post not in tl['posts']:
        print('Error: post %s not in timeline of user %s' %(
            post, user,
        ))
    else:
        payload = {
            'user_id': user,
            'post_id': post,
            'send_unix_ms': int(time.time() * 1000),
        }
        r = requests.post(upvote_url, json=payload)
        return json.loads(r.text)['post_id']

# --- test save
start_unix_ms = int(time.time()*1000)
print(start_unix_ms)
# tester-0 saves a posts
post_id = save_post(user=make_user(0), num_img=1)
time.sleep(0.5)


print('########################## Upvotes ##########################')
print('%s upvotes post %s' %(make_user(0), post_id))
tl = read_tl(user=make_user(0), earl_unix_ms=start_unix_ms, num_posts=10, user_tl=False)
upvote_tl_post(tl=tl, user=make_user(0), post=post_id)

print('%s upvotes post %s' %(make_user(2), post_id))
tl = read_tl(user=make_user(2), earl_unix_ms=start_unix_ms, num_posts=10, user_tl=False)
upvote_tl_post(tl=tl, user=make_user(2), post=post_id)

print('########################## user timeline ##########################')
# tester-0 should see 3 posts
tl = read_tl(user=make_user(0), earl_unix_ms=start_unix_ms, num_posts=10, user_tl=True)
print('# Reading post %s' %post_id)
print_tl_post(tl, post_id, make_user(0))





