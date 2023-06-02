import requests
import time
import threading
import random
import argparse

parser = argparse.ArgumentParser()
parser.add_argument('--c', dest='concurrency', type=int, default=1)
args = parser.parse_args()
concurrency = args.concurrency

random.seed(time.time())
all_texts = [
    'Pinocchio is next to you!',
    'Who is the clown?',
    'Clown Donald\'s vegetable has exploded!'
]
post_store_url = 'http://localhost:31998/v1.0/state/post-store-test/'

def sentiment(post_id, text):
    global dest_pubsub
    global dest_topic
    service_url = 'http://localhost:31998/v1.0/publish/sentiment-pubsub/sentiment'
    ts = int(round(time.time() * 1000))
    payload = {
        'post_id': post_id,
        'text': text,
        'send_unix_ms': ts, 
        'client_unix_ms': ts,
    }
    r = requests.post(service_url, json=payload)
    print(r.text)

def make_post_id(user: str, ts: int):
    return "%s*%d" %(user, ts)

def meta_key(post_id: str):
    return post_id + "-me"

threads = []
posts = []
for i in range(0, concurrency):
    pid = make_post_id('snow', i * 1000 + 500000)
    posts.append(pid)
    t = threading.Thread(
        target=sentiment, 
        kwargs={
            'post_id': pid,
            'text': random.choice(all_texts),
        }
    )
    threads.append(t)
    t.start()

for t in threads:
    t.join()