import sys
import time
import random
import json
import logging
import ffmpeg
from pathlib import Path
import argparse
from dapr.clients import DaprClient
from dapr.clients.grpc._state import StateItem, StateOptions, Consistency, Concurrency
from PIL import Image
import io
from typing import List

logging.basicConfig(level=logging.INFO)
parser = argparse.ArgumentParser()
parser.add_argument('--s', dest='save_video', action='store_true')
args = parser.parse_args()
save_video = args.save_video

random.seed(time.time())

pubsub_name = 'video-pubsub'
scale_topic_name = 'scale'
thumbnail_topic_name = 'thumbnail'
video_store = 'video-store-test'

def thumbnailId(video_id: str):
    if '.' in video_id:
        img_id = video_id.rsplit('.', 1)[0]
        if img_id == '':
            img_id = video_id
    else:
        img_id = video_id
    return "tn-%s.jpeg" %img_id

# save states
videos = ['earth_1920.avi', 'empty.avi']
video_dur = {}

if save_video:
    with DaprClient() as d:
        for video in videos:
            # check video info
            if 'empty' not in video:
                probe = ffmpeg.probe('video/' + video)
                video_dur[video] = float(probe['format']['duration'])
            else:
                video_dur[video] = 0
            with open('video/' + video, 'rb') as f:
                video_data = f.read()
                d.save_state(
                    store_name=video_store, 
                    key=video, 
                    value=video_data,
                    # options=StateOptions(consistency=Consistency.strong),
                    )
                print('%s saved' %video)

time.sleep(5)
def thumbnail_video(video_id: str):
    with DaprClient() as d:
        req_data = {
            'video_id': video_id,
            'duration': video_dur[video],
            'send_unix_ms': int(time.time() * 1000),
            'client_unix_ms': int(time.time() * 1000),
        }
        resp = d.publish_event(
            pubsub_name=pubsub_name,
            topic_name=thumbnail_topic_name,
            data=json.dumps(req_data),
            data_content_type='application/json',
        )
        print(resp)

for v in videos:
    thumbnail_video(v)
    time.sleep(10)
    tn_id = thumbnailId(v)
    print(tn_id)
    with DaprClient() as d:
        state = d.get_state(
            store_name=video_store, 
            key=tn_id, 
        )
        if len(state.data) != 0:
            pil_img = Image.open(io.BytesIO(state.data))
            print(pil_img)
        else:
            print('Non-existing')
        print('####################################')