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
# video_path = Path(__file__).parent.resolve() / '..' / 'video'
video_path = Path(__file__).parent.resolve() / '..' / '..' / 'workload' / 'vpipe' / 'video'
sys.path.append(str(video_path))

# read videos and encode as base64
videos = {}
b64_videos = {}
for vid in os.listdir(str(video_path)):
    if 'empty' in vid:
        continue
    with open(str(video_path / vid), 'rb') as f:
        print(vid)
        videos[vid] = f.read()
        b64_videos[vid] = base64.b64encode(videos[vid])
        base64.b64decode(b64_videos[vid])
        print(len(b64_videos[vid]))
print('video encoded')
test_dates = [
    '2022-01-01',
    '2022-01-02',
    '2022-01-03',
    '2022-01-04',
    '2022-01-05',
]

def make_user(i: int):
    return 'tester-' + str(i)

def upload_video(user: str, video_b64: str, desc: str, dt: str):
    unix_ms = int(time.time() * 1000)
    payload = {
        'user': user,
        'video_b64': video_b64,
        'description': desc,
        'date': dt,
        'send_unix_ms': int(time.time() * 1000),
    }
    r = requests.post(upload_url, json=payload)
    print(r.text)
    return json.loads(r.text)['video_id']

def get_video(video: str, res: str):
    payload = {
        'video': video,
        'resolution': res,
        'send_unix_ms': int(time.time() * 1000),
    }
    r = requests.get(video_url, json=payload)
    if len(r.text) < 2000:
        print(r.text)
    # print(r.text)
    return json.loads(r.text)['data']

def get_info(videos: List[str]):
    payload = {
        'videos': videos,
        'send_unix_ms': int(time.time() * 1000),
    }
    r = requests.get(info_url, json=payload)
    # print(r.text)
    return json.loads(r.text)['video_info']

def print_info(vid: str ,data: dict):
    info = data['info']
    thumbnail = data['thumbnail']
    print('---- video: %s' %vid)
    print('-------- thumbnail size: %d' %len(thumbnail))
    print('-------- views: %d' %info['views'])
    print('-------- score: %.3f' %info['score'])
    print('-------- num: %d' %info['num'])
    print('-------- user: %s' %info['meta']['user_id'])
    print('-------- reso: %s' %','.join(info['meta']['resolutions']))
    print('-------- dur: %.1fs' %info['meta']['duration'])
    print('-------- desc: %s' %info['meta']['description'])
    print('-------- date: %s' %info['meta']['date'])
    print('------------------------------')

def print_info_resp(video_info: dict):
    for vid in video_info:
        print_info(vid, video_info[vid])

all_vids = []
v_to_vids = {}
print('#------------ upload-video -------------#')
for v in b64_videos:
    # --- test upload
    unix_ms = int(time.time() * 1000)
    vid = upload_video(
        user=make_user(0), 
        video_b64=b64_videos[v], 
        # video_b64=b64_videos['SampleVideo_1280x720_2mb.mp4'],
        desc='%s shouts out at %d: Fakers get out of academia! --- from %s' %(
            make_user(0), unix_ms, v), 
        dt=random.choice(test_dates))
    print('%s uploaded' %v)
    all_vids.append(vid)
    print(vid)
    v_to_vids[v] = vid
    time.sleep(5)

for v in v_to_vids:
    print('%s -> %s' %(v, v_to_vids[v]))

time.sleep(15)
# --- test get info
print('#------------ get-video-info -------------#')
info_resp = get_info(
    videos=all_vids,
)
print_info_resp(info_resp)

time.sleep(60)
# --- test get video
print('#------------ get-video -------------#')
for vid in all_vids:
    all_res = info_resp[vid]['info']['meta']['resolutions']
    for res in all_res:
        d = get_video(
            video=vid,
            res=res,
        )
        print('vid: %s, resolution=%s, size=%d' %(vid, res, len(d)))
