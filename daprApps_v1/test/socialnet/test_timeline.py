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
        print('---- upvotes: %s' %(','.join([post['upvotes']])))

def test_tl_img(tl):
    print('-- num_posts: %d' %len(tl['posts']))
    for post_id in tl['posts']:
        if tl['posts'][post_id]['content']['images'] != None:
            print('-- images in post: %s' %post_id)
            for img in tl['posts'][post_id]['content']['images']:
                payload = {
                    'image': img,
                    'send_unix_ms': int(time.time() * 1000),
                }
                r = requests.get(image_url, json=payload)
                print(len(r.text))
                pil_img = Image.open(io.BytesIO(base64.b64decode(r.text)))
                print('----', pil_img)
    print('########################')

# --- test save
start_unix_ms = int(time.time()*1000)
print(start_unix_ms)
# tester-0 saves 3 posts
save_post(user=make_user(0), num_img=1)
time.sleep(0.5)
save_post(user=make_user(0), num_img=1)
time.sleep(0.5)
save_post(user=make_user(0), num_img=1)
time.sleep(0.5)
# tester-1 saves 2 post
save_post(user=make_user(1), num_img=1)
time.sleep(0.5)
save_post(user=make_user(1), num_img=1)
time.sleep(0.5)
# tester-2 saves 1 post
save_post(user=make_user(2), num_img=1)
time.sleep(15)

print('########################## home timeline ##########################')
# tester-0 should see 6 posts, 3 from tester-0, 2 from tester-1 and 1 from tester-2
tl = read_tl(user=make_user(0), earl_unix_ms=start_unix_ms, num_posts=10, user_tl=False)
print('# home timeline of %s should have 6 posts' %make_user(0))
print_tl(tl)
# tester-1 should see 3 posts, 2 from tester-1 and 1 from tester-2
tl = read_tl(user=make_user(1), earl_unix_ms=start_unix_ms, num_posts=10, user_tl=False)
print('# home timeline of %s should have 3 posts' %make_user(1))
print_tl(tl)
# tester-2 should see 4 posts, 3 from tester-0 and 1 from tester-2
tl = read_tl(user=make_user(2), earl_unix_ms=start_unix_ms, num_posts=10, user_tl=False)
print('# home timeline of %s should have 4 post' %make_user(2))
print_tl(tl)

print('########################## user timeline ##########################')
# tester-0 should see 3 posts
tl = read_tl(user=make_user(0), earl_unix_ms=start_unix_ms, num_posts=10, user_tl=True)
print('# user timeline of %s should have 3 posts' %make_user(0))
print_tl(tl)
# tester-1 should see 2 posts
tl = read_tl(user=make_user(1), earl_unix_ms=start_unix_ms, num_posts=10, user_tl=True)
print('# user timeline of %s should have 2 posts' %make_user(1))
print_tl(tl)
# tester-2 should see 1 post
tl = read_tl(user=make_user(2), earl_unix_ms=start_unix_ms, num_posts=10, user_tl=True)
print('# user timeline of %s should have 1 post' %make_user(2))
print_tl(tl)

print('########################## Read image ##########################')
tl = read_tl(user=make_user(0), earl_unix_ms=start_unix_ms, num_posts=10, user_tl=False)
test_tl_img(tl)

print('########################## Delete post ##########################')
tl = read_tl(user=make_user(0), earl_unix_ms=start_unix_ms, num_posts=10, user_tl=True)
last_post = tl_last_post(tl)
print('## del post %s from %s' %(last_post, make_user(0)))
# delete the last post from user(0)
del_post(user=make_user(0), post=last_post)
time.sleep(1)
# tester-0 should see 5 posts, 2 from tester-0, 2 from tester-1 and 1 from tester-2
tl = read_tl(user=make_user(0), earl_unix_ms=start_unix_ms, num_posts=10, user_tl=False)
print('# home timeline of %s should have 5 posts' %make_user(0))
print_tl(tl)
# tester-1 should see 3 posts, 2 from tester-1 and 1 from tester-2
tl = read_tl(user=make_user(1), earl_unix_ms=start_unix_ms, num_posts=10, user_tl=False)
print('# home timeline of %s should have 3 posts' %make_user(1))
print_tl(tl)
# tester-2 should see 3 posts, 2 from tester-0 and 1 from tester-2
tl = read_tl(user=make_user(2), earl_unix_ms=start_unix_ms, num_posts=10, user_tl=False)
print('# home timeline of %s should have 3 post' %make_user(2))
print_tl(tl)
# tester-0 should see 2 posts
tl = read_tl(user=make_user(0), earl_unix_ms=start_unix_ms, num_posts=10, user_tl=True)
print('# user timeline of %s should have 2 posts' %make_user(0))
print_tl(tl)





