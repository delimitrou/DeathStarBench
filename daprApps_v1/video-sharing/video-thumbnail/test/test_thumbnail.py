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

util_path = Path(__file__).parent.resolve() / 'dapr_client'
sys.path.append(str(util_path))
from dapr_client import MyDaprClient

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
    return "tn-%s.jpeg" %video_id

# save states
# save states
# videos = [
#     'earth_1920.avi',
#     'empty.avi',
#     'sample-5s.mp4',
#     'SampleVideo_1280x720_1mb.mp4',
#     'SampleVideo_1280x720_2mb.mp4',
#     'SampleVideo_720x480_1mb.mp4',
#     'SampleVideo_720x480_2mb.mp4',
#     # 'short_sample_1.mp4',
#     # 'short_sample_2.mp4',
#     # 'short_sample_3.mp4',
#     # 'short_sample_4.mp4',
#     # 'short_sample_5.mp4',
#     # 'short_sample_6.mp4',
#     # 'short_sample_7.mp4',
# ]
videos = [
    # 'short_sample_7.mp4',
    'short_sample_6.mp4',
    'short_sample_5.mp4',
    'short_sample_4.mp4',
    'short_sample_3.mp4',
    'short_sample_2.mp4',
    'short_sample_1.mp4',
    'sample-5s.mp4',
    'sample-10s.mp4',
    'SampleVideo_720x480_5mb.mp4',
    'SampleVideo_720x480_2mb.mp4',
    'SampleVideo_720x480_1mb.mp4',
    'SampleVideo_1280x720_5mb.mp4',
    'SampleVideo_1280x720_2mb.mp4',
    'SampleVideo_1280x720_1mb.mp4',
    # 'SampleVideo_1280x720_10mb.mp4',
    'earth_1920.avi',
    'empty.avi',
]
video_dur = {}

MAX_MESSAGE_LENGTH=33554432
grpcOptions = [
    ('grpc.max_send_message_length', MAX_MESSAGE_LENGTH),
    ('grpc.max_receive_message_length', MAX_MESSAGE_LENGTH),
]

ts = int(time.time() * 1000)

all_vids = []
with MyDaprClient(options=grpcOptions) as d:
    for video in videos:
        # check video info
        vid = str(ts) + '-' + video
        all_vids.append(vid)
        if 'empty' not in video:
            probe = ffmpeg.probe('video/' + video)
            video_dur[vid] = float(probe['format']['duration'])
        else:
            video_dur[vid] = 0
        if save_video:
            with open('video/' + video, 'rb') as f:
                video_data = f.read()
                d.save_state(
                    store_name=video_store, 
                    key=vid, 
                    value=video_data,
                    # options=StateOptions(consistency=Consistency.strong),
                    )
                print('%s saved' %vid)

time.sleep(5)
def thumbnail_video(video_id: str):
    with MyDaprClient(options=grpcOptions) as d:
        req_data = {
            'video_id': video_id,
            'data_id': video_id,
            'duration': video_dur[video_id],
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

for v in all_vids:
    thumbnail_video(v)
    time.sleep(10)
    tn_id = thumbnailId(v)
    print(tn_id)
    with MyDaprClient(options=grpcOptions) as d:
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