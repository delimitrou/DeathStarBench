import requests
import time
from datetime import datetime
import json
import random
from typing import Optional, Union, Callable, List

# dapr-trending
trend_url = 'http://localhost:31990/v1.0/invoke/dapr-trending/method/get'
# dapr-video-info
info_upload_url = 'http://localhost:31991/v1.0/invoke/dapr-video-info/method/upload'
rate_url = 'http://localhost:31991/v1.0/invoke/dapr-video-info/method/rate'
view_url = 'http://localhost:31991/v1.0/invoke/dapr-video-info/method/view'
# dates url
dates_upload_url = 'http://localhost:31994/v1.0/invoke/dapr-dates/method/upload'
dates_get_url = 'http://localhost:31994/v1.0/invoke/dapr-dates/method/get'

# helper functions
def make_date():
    now = datetime.now()
    return now.strftime('%Y-%m-%d')

def make_date_upload(video_id: str, dt: str):
    return {
        'video_id': video_id,
        'date': dt,
        'send_unix_ms':int(time.time() * 1000),
    }

def make_dates_get(dts: List[str]):
    return {
        'dates': dts,
        'send_unix_ms':int(time.time() * 1000),
    }

def show_dates_get_resp(respjson):
    respdata = json.loads(respjson)
    print(respdata)
    for dt in respdata['videos']:
        print('-- date: %s' %(dt))
        print('------ videos: %s' % ', '.join(respdata['videos'][dt]))

def store_date(d: str, videos: List[str]):
    # data = {
    #     'videos': videos,
    # }
    for v in videos:
        req = make_date_upload(v, d)
        r = requests.post(dates_upload_url, json=req)
        print(r.text)

def get_dates(ds: List[str]):
    req = make_dates_get(ds)
    r = requests.get(dates_get_url, json=req)
    show_dates_get_resp(r.text)

def make_info_upload(video_id: str, user_id: str, reso: List[str],
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
        'send_unix_ms': int(time.time() * 1000),
    }

def make_view(video_id: str):
    return {
        'video_id': video_id,
        'send_unix_ms': int(time.time() * 1000),
    }

def make_trending(start_d: str, end_d: str):
    return {
        'start_date': start_d,
	    'end_date':end_d,
	    'send_unix_ms': int(time.time() * 1000),
    }

# variables
# user_id = 'test-ur-user-0'
# video_id = 'test-ur-video-0'
user_id = 'Justice'
video_id = 'The fallen ca'

all_dates = [
    '2023-03-01',
    '2023-03-02',
    '2023-03-03',
    '2023-03-04',
    '2023-03-05',
]
all_videos = {}
all_videos['2023-03-01'] = ['test-tr-video-0', 'test-tr-video-1'] 
all_videos['2023-03-02'] = ['test-tr-video-2', 'test-tr-video-3'] 
all_videos['2023-03-03'] = ['test-tr-video-4', 'test-tr-video-5'] 
all_videos['2023-03-04'] = ['test-tr-video-6', 'test-tr-video-7'] 
all_videos['2023-03-05'] = ['test-tr-video-8', 'test-tr-video-9'] 

# store dates
for d in all_videos:
    store_date(d, all_videos[d])
print('------- get videos of each date -------')
get_dates(all_dates)
print('---------------------')

uid = 0
# upload videos
for d in all_videos:
    for v in all_videos[d]:
        upload_req = make_info_upload(
            video_id=v, 
            user_id='Integrity-' + str(uid), 
            reso=['1080p'],
            dur=10.0,
            desc='The community has fallen!', 
        )
        uid += 1
        r = requests.post(info_upload_url, json=upload_req)
        print(r.text)

for d in ['2023-03-05', '2023-03-02', '2023-03-01']:
    for v in all_videos[d]:
        for i in range(0, random.randint(5, 20)):
            rate_req = make_rate(
                video_id=v, 
                change=False, 
                score=random.randint(1, 5),
                ori_score=0.0)
            r = requests.post(rate_url, json=rate_req)
            print(r.text)
            view_req = make_view(video_id=v)
            r = requests.post(view_url, json=view_req)
            print(r.text)

# only rate videos in '2023-03-04' once
for v in all_videos['2023-03-04']:
    rate_req = make_rate(
        video_id=v, 
        change=False, 
        score=random.randint(1, 5),
        ori_score=0.0)
    r = requests.post(rate_url, json=rate_req)
    print(r.text)
    view_req = make_view(video_id=v)
    r = requests.post(view_url, json=view_req)
    print(r.text)

# only view videos in '2023-03-03'
for v in all_videos['2023-03-03']:
    view_req = make_view(video_id=v)
    r = requests.post(view_url, json=view_req)
    print(r.text)

tr_req = make_trending('2023-03-01', '2023-03-05')
r = requests.get(trend_url, json=tr_req)
print(r.text)


