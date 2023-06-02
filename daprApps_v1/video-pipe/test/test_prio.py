import requests
import time
from pathlib import Path
import sys
import os

# events
meta_url = 'http://localhost:31981/v1.0/publish/vpipe-events/meta'
scene_url = 'http://localhost:31981/v1.0/publish/vpipe-events/scene'
face_url = 'http://localhost:31981/v1.0/publish/vpipe-events/face'
# stores
video_store_url = 'http://localhost:31981/v1.0/state/vpipe-video-store/'
image_store_url = 'http://localhost:31981/v1.0/state/vpipe-image-store/'

video_dir = Path(__file__).parent.resolve().parent.parent.resolve() / 'workload' / 'vpipe' / 'locust' / 'b64_video'
video_ids = []
for fn in os.listdir(video_dir):
    video_ids.append(fn)

def make_meta_req(video_id: str, req_id: str, priority: int):
    assert priority == 1 or priority == 2
    return {
        'video_id': video_id,
        'req_id': req_id,
        'send_unix_ms': int(time.time() * 1000),
        'priority': priority,
    }

payload = make_meta_req(
    # video_id=video_ids[0], 
    video_id='b64_human10.mp4',
    req_id='req-0', 
    priority=1
)
print(payload)
r = requests.post(meta_url, json=payload)
print(r)
print(r.text)

payload = make_meta_req(
    video_id=video_ids[1], 
    req_id='req-1', 
    priority=1
)
print(payload)
r = requests.post(meta_url, json=payload)
print(r)
print(r.text)

payload = make_meta_req(
    video_id=video_ids[2], 
    req_id='req-2', 
    priority=2
)
print(payload)
r = requests.post(meta_url, json=payload)
print(r)
print(r.text)


