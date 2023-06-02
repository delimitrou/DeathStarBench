import requests
import time
import threading
import argparse
import json
import random
import numpy as np
from typing import Optional, Union, Callable, List

parser = argparse.ArgumentParser()
parser.add_argument('--c', dest='concurrency', type=int, default=10)
args = parser.parse_args()
concurrency = args.concurrency

get_url = 'http://localhost:31992/v1.0/invoke/dapr-user-rating/method/get'
rate_url = 'http://localhost:31992/v1.0/invoke/dapr-user-rating/method/rate'

# state store url
state_url = 'http://localhost:31991/v1.0/state/user-rating-store-test/'

# helper functions
def make_rate(user_id: str, video_id: str, 
        score: float, comment: str):
    return {
        'user_id': user_id,
        'video_id': video_id,
        'score': score,
        'comment': comment,
        'send_unix_ms':int(time.time() * 1000),
    }

def make_get(user_id: str, video_id: str):
    return {
        'user_id': user_id,
        'video_id': video_id,
        'send_unix_ms':int(time.time() * 1000),
    }

def show_get_resp(user_id, video_id, respjson):
    respdata = json.loads(respjson)
    print('-- user_id: %s, video_id: %s' %(
        user_id, video_id))
    print('------ exist: %s' %respdata['exist'])
    print('------ score: %.3f' %respdata['score'])
    print('------ comment: %s' %respdata['comment'])

def show_rate_resp(respjson):
    respdata = json.loads(respjson)
    print('------ exist: %s' %respdata['exist'])
    print('------ ori_score: %.3f' %respdata['ori_score'])

def update_rate(user_id: str, video_id: str, score: float, comment: str):
    rate_req = make_rate(
        user_id=user_id, 
        video_id=video_id, 
        score=score, 
        comment=comment, 
    )
    r = requests.post(rate_url, json=rate_req)

user_id = 'Justice'
video_id = 'No justice in this community ep-'

threads = []
videos = [video_id + str(i) for i in range(0, concurrency)]
comments = ['So true for this community, ep-' + str(i) for i in range(0, concurrency)]
for i in range(0, concurrency):
    t = threading.Thread(
        target= update_rate,
        kwargs={
            'user_id': user_id,
            'video_id': videos[i],
            'score': 5.0,
            'comment': comments[i],
        }
    )
    threads.append(t)
    t.start()

for t in threads:
    t.join()

print("------ get rate ------")
for v in videos:
    read_req = make_get(user_id=user_id, video_id=v)
    r = requests.get(get_url, json=read_req)
    show_get_resp(user_id, v, r.text)
