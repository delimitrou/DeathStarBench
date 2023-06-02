import requests
import time
from datetime import datetime
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

upload_url = 'http://localhost:31991/v1.0/invoke/dapr-video-info/method/upload'
rate_url = 'http://localhost:31991/v1.0/invoke/dapr-video-info/method/rate'
view_url = 'http://localhost:31991/v1.0/invoke/dapr-video-info/method/view'
info_url = 'http://localhost:31991/v1.0/invoke/dapr-video-info/method/info'
# state store url
state_url = 'http://localhost:31991/v1.0/state/info-store-test/'

# helper functions
def make_date():
    now = datetime.now()
    return now.strftime('%Y-%m-%d')

# helper functions
def make_upload(video_id: str, user_id: str, reso: List[str],
        dur: float, desc: str):
    return {
        'video_id': video_id,
        'user_id': user_id,
        'resolutions': reso,
        'duration': dur,
        'description': desc,
        'date': make_date(),
        'send_unix_ms': int(time.time() * 1000),
    }

def make_rate(video_id: str, change: bool, score: float,
        ori_score: float):
    return {
        'video_id': video_id,
        'change': change,
        'score': score,
        'ori_score': ori_score,
        'send_unix_ms':int(time.time() * 1000),
    }

def make_view(video_id: str):
    return {
        'video_id': video_id,
        'send_unix_ms':int(time.time() * 1000),
    }

def make_info(video_ids: List[str]):
    return {
        'video_ids': video_ids,
        'send_unix_ms': int(time.time() * 1000),
    }

def show_info(infojson: str):
    infodata= json.loads(infojson)
    info = infodata['info']
    for video_id in info:
        video = info[video_id]
        print('-- video_id:', video_id)
        print('------ user:', video['meta']['user_id'])
        print('------ reso:', video['meta']['resolutions'])
        print('------ dur:',  video['meta']['duration'])
        print('------ desc:', video['meta']['description'])
        print('------ date:', video['meta']['date'])
        print('------ rate_num:', video['rating']['num'])
        print('------ rate_score:', video['rating']['score'])
        print('------ rate_score_sq:', video['rating']['score_sq'])
        print('------ views:', video['views'])

def update_view(video_id: str):
    view_req = make_view(
        video_id=video_id)
    r = requests.post(view_url, json=view_req)

video_id = 'cncr-view-video-x'
print("------ upload video ------")
upload_req = make_upload(
    video_id=video_id, 
    user_id='Integrity', 
    reso=['1080p'],
    dur=10.0,
    desc='The community has fallen!', 
)
r = requests.post(upload_url, json=upload_req)
print(r.text)

print("------ video info ------")
read_req = make_info(video_ids=[video_id])
r = requests.get(info_url, json=read_req)
show_info(r.text)

threads = []
for i in range(0, concurrency):
    t = threading.Thread(
        target= update_view,
        kwargs={
            'video_id': video_id,
        }
    )
    threads.append(t)
    t.start()

for t in threads:
    t.join()

print("------ video info, expected views=%d ------" %concurrency)
read_req = make_info(video_ids=[video_id])
r = requests.get(info_url, json=read_req)
show_info(r.text)