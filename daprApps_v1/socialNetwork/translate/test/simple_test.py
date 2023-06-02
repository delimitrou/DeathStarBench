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

def transl(text):
    service_url = 'http://localhost:31999/v1.0/invoke/dapr-transl-en-to-de/method/en_to_de'
    payload = {'send_unix_ms': int(round(time.time() * 1000)), 
        'text': text}
    r = requests.post(service_url, json=payload)
    print(r.text)

concurrency = 25
threads = []
for i in range(0, concurrency):
    t = threading.Thread(target=transl, kwargs={
        'text': random.choice(all_texts)
    })
    threads.append(t)
    t.start()

for t in threads:
    t.join()