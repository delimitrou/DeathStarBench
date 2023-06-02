import os
import sys
import requests
import time
import json
import base64
from pathlib import Path
import random
from PIL import Image
from typing import List
import io

random.seed(int(time.time()))

upload_url = 'http://localhost:31985/v1.0/invoke/dapr-video-frontend/method/upload'
info_url   = 'http://localhost:31985/v1.0/invoke/dapr-video-frontend/method/info'
video_url = 'http://localhost:31985/v1.0/invoke/dapr-video-frontend/method/video'
rate_url = 'http://localhost:31985/v1.0/invoke/dapr-video-frontend/method/rate'
get_rate_url = 'http://localhost:31985/v1.0/invoke/dapr-video-frontend/method/get_rate'
trending_url = 'http://localhost:31986/v1.0/invoke/dapr-trending/method/get'

from pathlib import Path
video_path = Path(__file__).parent.resolve() / '..' / 'video'
sys.path.append(str(video_path))

# read videos and encode as base64
test_dates = [
    '2022-01-01',
    '2022-01-02',
    '2022-01-03',
    '2022-01-04',
    '2022-01-05',
]

def make_user(i: int):
    return 'tester-' + str(i)

def get_trending(start_dt: str, end_dt: str):
    payload = {
        'start_date': start_dt,
        'end_date': end_dt,
        'send_unix_ms': int(time.time() * 1000),
    }
    r = requests.get(trending_url, json=payload)
    # print(r.text)
    return json.loads(r.text)['videos']

def upload_rate(user_id: str, video_id: str, comment: str,
        score: float):
    payload = {
        'user': user_id,
        'video': video_id,
        'comment': comment,
        'score': score,
        'send_unix_ms': int(time.time() * 1000),
    }
    r = requests.post(rate_url, json=payload)
    # print(r.text)
    return json.loads(r.text)['video_id']

def get_info(videos: List[str]):
    payload = {
        'videos': videos,
        'send_unix_ms': int(time.time() * 1000),
    }
    r = requests.get(info_url, json=payload)
    # print(r.text)
    return json.loads(r.text)['video_info']

def get_rate(user_id: str, video_id: str):
    payload = {
        'user': user_id,
        'video': video_id,
        'send_unix_ms': int(time.time() * 1000),
    }
    r = requests.get(get_rate_url, json=payload)
    # print(r.text)
    return json.loads(r.text)

def print_info(vid: str, data: dict):
    info = data['info']
    thumbnail = data['thumbnail']
    print('---- video: %s' %vid)
    # print('-------- thumbnail size: %d' %len(thumbnail))
    print('-------- views: %d' %info['views'])
    print('-------- score: %.3f' %info['score'])
    print('-------- num: %d' %info['num'])
    # print('-------- user: %s' %info['meta']['user_id'])
    # print('-------- reso: %s' %','.join(info['meta']['resolutions']))
    # print('-------- dur: %.1fs' %info['meta']['duration'])
    # print('-------- desc: %s' %info['meta']['description'])
    # print('-------- date: %s' %info['meta']['date'])
    # print('------------------------------')

def print_info_resp(video_info: dict):
    for vid in video_info:
        print_info(vid, video_info[vid])

def print_trending(vids):
    print('---- trending: ')
    for v in vids:
        print('-------- %s' %v)
    print('------------------------------')

def print_user_rate(user: str, vid: str, rate: dict):
    print('---- rate video: %s, user: %s' %(vid, user))
    # print('-------- thumbnail size: %d' %len(thumbnail))
    print('-------- exist: %s' %str(rate['exist']))
    print('-------- score: %.1f' %rate['score'])
    print('-------- comment: %s' %rate['comment'])

# get videos 

vids = get_trending(test_dates[0], test_dates[-1])
print_trending(vids)
# --- test get info
info_resp = get_info(
    videos=vids,
)
print_info_resp(info_resp)
print('----------- rate videos -----------')
user_to_video = {}
for i in range(0, 200):
    u = 'tester-' + str(random.randint(0, 100))
    v = random.choice(vids)
    if u not in user_to_video:
        user_to_video[u] = []
    if v not in user_to_video[u]:
        user_to_video[u] += [v]
    upload_rate(
        user_id=u, 
        video_id=v, 
        comment='Disappointed. Such disgrace -- %d' %i,
        score=float(random.randint(1,5)))
    # time.sleep(0.01)

print('----------- after rating videos -----------')
vids = get_trending(test_dates[0], test_dates[-1])
print_trending(vids)
# --- test get info
info_resp = get_info(
    videos=vids,
)
print_info_resp(info_resp)

print('----------- get user rating -----------')
for u in user_to_video:
    for v in user_to_video[u]:
        r = get_rate(u, v)
        print_user_rate(u, v, r)
