import sys
import os
import time
import threading
import random
import json
import logging
from pathlib import Path
import argparse
from dapr.clients import DaprClient
from dapr.clients.grpc._state import StateItem, StateOptions, Consistency, Concurrency
from PIL import Image
import io
from typing import List

logging.basicConfig(level=logging.INFO)
parser = argparse.ArgumentParser()
parser.add_argument('--c', dest='concurrency', type=int, default=1)
parser.add_argument('--s', dest='save_image', action='store_true')
args = parser.parse_args()
concurrency = args.concurrency
save_image = args.save_image

random.seed(time.time())

pubsub_name = 'object-detect-pubsub'
topic_name = 'object-detect'

img_store = os.getenv('IMAGE_STORE', 'image-store-test')
post_store = os.getenv('POST_STORE', 'post-store-test')

# save states
images = [
    'panda2.jpg',  
    'panda.jpeg',  
    'shiba2.jpg',  
    'shiba.jpg',
    ]
if save_image:
    with DaprClient() as d:
        for img in images:
            with open('images/' + img, 'rb') as f:
                img_data = f.read()
                pil_img = Image.open(io.BytesIO(img_data))
                print(pil_img)
                # logging.info(img)
                logging.info('%s, len=%d, size=%d' %(img, len(img_data), sys.getsizeof(img_data)))
                d.save_state(
                    store_name=img_store, 
                    key=img, 
                    value=img_data,
                    # options=StateOptions(consistency=Consistency.strong),
                    )
                print('%s saved' %img)

def make_post_id(user: str, ts: int):
    return "%s*%d" %(user, ts)

def meta_key(post_id: str):
    return post_id + "-me"

# time.sleep(5)
def object_detect(post_id: str, images: List[str]):
    with DaprClient() as d:
        ts = int(time.time() * 1000)
        req_data = {
            'post_id': post_id,
            'pubsub_name': pubsub_name,
            'topic_name': topic_name,
            'images': images,
            'send_unix_ms': ts,
            'client_unix_ms': ts,
        }
        resp = d.publish_event(
            pubsub_name=pubsub_name,
            topic_name=topic_name,
            data=json.dumps(req_data),
            data_content_type='application/json',
        )
        print(resp)

threads = []
posts = []
for i in range(0, concurrency):
    n = random.randint(1, len(images))
    l = list(images)
    random.shuffle(l)
    pid = make_post_id('snow', i * 1000 + 500000)
    posts.append(pid)
    t = threading.Thread(
        target=object_detect, 
        kwargs={
            'post_id': pid,
            'images': l[:n],
        }
    )
    threads.append(t)
    t.start()

for t in threads:
    t.join()

print('--------- post ids ---------')
print(posts)

time.sleep(0.5)
print('--------- check post meta from post store --------')
with DaprClient() as d:
    for p in posts:
        print('-------------')
        print(p)
        state = d.get_state(
            store_name=post_store, 
            key=meta_key(p), 
        )
        print(json.loads(state.data))
