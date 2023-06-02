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

video_dir = Path(__file__).parent.resolve().parent.parent.resolve() / 'workload' / 'vpipe'  / 'locust' / 'b64_video'
video_ids = []
for fn in os.listdir(video_dir):
    video_ids.append(fn)

image_ids = [
    'req-0_01.jpg',
    'req-0_02.jpg',
    'req-0_03.jpg',
    'req-0_04.jpg',
    'req-0_05.jpg',
    'req-0_06.jpg',
]

for img in image_ids:
    print(img)
    r = requests.get(image_store_url + img)
    print(r)
    print(len(r.text))


