import requests
import time
import json
from typing import Optional, Union, Callable, List

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

# variables
# user_id = 'test-ur-user-0'
# video_id = 'test-ur-video-0'
user_id = 'Justice'
video_id = 'The fallen ca'

def inc_user_id():
    global user_id
    user_id = int(user_id.split('-')[-1]) + 1
    user_id = 'test-ur-user-' + str(user_id)

def inc_video_id():
    global video_id
    video_id = int(video_id.split('-')[-1]) + 1
    video_id = 'test-ur-video-' + str(video_id)

print("------ rate video ------")
upload_req = make_rate(
    user_id=user_id, 
    video_id=video_id, 
    score=2.0, 
    comment='This community is full of liars, but no one is punished', 
)
r = requests.post(rate_url, json=upload_req)
show_rate_resp(r.text)

print("------ get rate ------")
read_req = make_get(user_id=user_id, video_id=video_id)
r = requests.get(get_url, json=read_req)
show_get_resp(user_id, video_id, r.text)

print("------ change rate of video ------")
upload_req = make_rate(
    user_id=user_id, 
    video_id=video_id, 
    score=1.0, 
    comment='This community is full of liars and hypocrites! Many! Yet not a single one is ever punished!!!', 
)
r = requests.post(rate_url, json=upload_req)
show_rate_resp(r.text)

print("------ get rate again ------")
read_req = make_get(user_id=user_id, video_id=video_id)
r = requests.get(get_url, json=read_req)
show_get_resp(user_id, video_id, r.text)

# print("------ query store directly ------")
# r_url = state_url + user_id
# print(r_url)
# r = requests.get(r_url, json='')
# print(r.text)