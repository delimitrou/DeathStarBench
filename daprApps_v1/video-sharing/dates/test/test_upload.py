import requests
import time
from datetime import datetime
import json
from typing import Optional, Union, Callable, List

get_url = 'http://localhost:31994/v1.0/invoke/dapr-dates/method/get'
upload_url = 'http://localhost:31994/v1.0/invoke/dapr-dates/method/upload'

# state store url
state_url = 'http://localhost:31991/v1.0/state/date-store-test/'

# helper functions
def make_date():
    now = datetime.now()
    return now.strftime('%Y-%m-%d')

def make_upload(video_id: str, dt: str):
    return {
        'video_id': video_id,
        'date': dt,
        'send_unix_ms':int(time.time() * 1000),
    }

def make_get(dts: List[str]):
    return {
        'dates': dts,
        'send_unix_ms':int(time.time() * 1000),
    }

def show_get_resp(respjson):
    respdata = json.loads(respjson)
    print(respdata)
    for dt in respdata['videos']:
        print('-- date: %s' %(dt))
        print('------ videos: %s' % ', '.join(respdata['videos'][dt]))

# variables
# user_id = 'test-ur-user-0'
# video_id = 'test-ur-video-0'
video_id = 'The fallen ca community - 2017'
dt = '2021-02-12'

def inc_video_id():
    global video_id
    id = int(video_id.split('-')[-1]) + 1
    video_id = 'The fallen ca community - ' + str(id)

for i in range(0, 10):
    print("------ upload video ------")
    upload_req = make_upload(
        video_id=video_id, 
        dt=dt, 
    )
    r = requests.post(upload_url, json=upload_req)

    print("------ get video list ------")
    read_req = make_get(dts=[dt])
    r = requests.get(get_url, json=read_req)
    show_get_resp(r.text)
    inc_video_id()

dt2 = '2021-02-24'
for i in range(0, 5):
    print("------ upload video ------")
    upload_req = make_upload(
        video_id=video_id, 
        dt=dt2, 
    )
    r = requests.post(upload_url, json=upload_req)

    print("------ get video list ------")
    read_req = make_get(dts=[dt, dt2])
    r = requests.get(get_url, json=read_req)
    show_get_resp(r.text)
    inc_video_id()