import requests
import time
import random
import json

transl_url = 'http://localhost:31993/v1.0/invoke/dapr-transl-en-to-de/method/en_to_de'

random.seed(time.time())
all_texts = [
    'Pinocchio is next to you!',
    'Who is the clown?',
    'Clown Donald\'s vegetable has exploded!',
    'Fakers out of academia!',
    'Sorry for the last-minute cancellation, but YouKnowWho informed me a couple of students have come down with Covid (sorry to hear that please get well soon!).  I will reschedule this meeting.'
]

def transl(text):
    payload = {
        'send_unix_ms': int(round(time.time() * 1000)), 
        'text': text
    }
    r = requests.post(transl_url, json=payload)
    tr = json.loads(r.text)['translation']
    print('%s -> %s' %(text, tr))

for t in all_texts:
    transl(t)
    time.sleep(5)