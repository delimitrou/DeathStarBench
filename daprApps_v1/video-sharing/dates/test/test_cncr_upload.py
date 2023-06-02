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


def upload(video_id: str, dt: str):
    upload_req = make_upload(
        dt=dt,
        video_id=video_id,
    )
    r = requests.post(upload_url, json=upload_req)

user_id = 'Justice'
video_id = 'No justice in this community ep-'

dt = '2021-02-16'
threads = []
videos = [video_id + str(i) for i in range(0, concurrency)]
for i in range(0, concurrency):
    t = threading.Thread(
        target= upload,
        kwargs={
            'video_id': videos[i],
            'dt': dt,
        }
    )
    threads.append(t)
    t.start()

for t in threads:
    t.join()

print("------ get videos ------")
read_req = make_get(dts=[dt])
r = requests.get(get_url, json=read_req)
show_get_resp(r.text)
