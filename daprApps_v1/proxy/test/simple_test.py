import requests
import time
import threading
import random

random.seed(time.time())
all_texts = [
    'Pinocchio is next to you!',
    'Who is the clown?',
    'Clown Donald\'s vegetable has exploded!'
]

def forward(text):
    service_url = 'http://localhost:31789/v1.0/invoke/dapr-proxy/method/forward'
    cont = {
        'user_id': 'Integrity',
        'text': 'Where is the drone?',
        'images': [],
    }
    epoch = int(time.time()*1000)
    post_id = 'Integrity*' + str(epoch)
    payload = {
        'downstream': 'dapr-post',
        'method': 'save',
        'post_id': post_id,
        'content': cont,
        'send_unix_ms': epoch, 
    }
    r = requests.post(service_url, json=payload)
    print(r.text)

concurrency = 10
threads = []
for i in range(0, concurrency):
    t = threading.Thread(
        target=forward, 
        kwargs={
            'text': random.choice(all_texts)
        })
    threads.append(t)
    t.start()

for t in threads:
    t.join()