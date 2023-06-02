import sys
import time
import random
import json
import logging
import ffmpeg
from pathlib import Path
import argparse
# from dapr.clients import DaprClient
from dapr.clients.grpc._state import StateItem, StateOptions, Consistency, Concurrency
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

def scaledVideoId(video_id: str, width: int):
    if width == 640:
        return '480p-%s' %(video_id)
    elif width == 1280:
        return '720p-%s' %(video_id)
    elif width == 1920:
        return '1080p-%s' %(video_id)
    else:
        print('Invalid video width %d' %width)
        return None

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
if save_video:
    with MyDaprClient(options=grpcOptions) as d:
        for video in videos:
            vid = str(ts) + '-' + video
            all_vids.append(vid)
            # check video info
            if 'empty' not in video:
                probe = ffmpeg.probe('video/' + video)
                video_dur[vid] = float(probe['format']['duration'])
            else:
                video_dur[vid] = 0
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
def scale_video(video_id: str, width: int, height: int):
    with MyDaprClient(options=grpcOptions) as d:
        req_data = {
            'video_id': video_id,
            'data_id': video_id,
            'width': width,
            'height': height,
            'send_unix_ms': int(time.time() * 1000),
            # 'client_unix_ms': int(time.time() * 1000),
        }
        resp = d.publish_event(
            pubsub_name=pubsub_name,
            topic_name=scale_topic_name,
            data=json.dumps(req_data),
            data_content_type='application/json',
        )
        print(resp)

threads = []
posts = []
for w, h in [[640, 480], [1280, 720]]:
    for v in all_vids:
        scale_video(v, w, h)
        time.sleep(30)
        scaled_vid = scaledVideoId(v, w)
        print(scaled_vid)
        with MyDaprClient(options=grpcOptions) as d:
            state = d.get_state(
                store_name=video_store, 
                key=scaled_vid, 
            )
            if len(state.data) != 0:
                scaled_fn = '%s-%s' %(int(time.time()*1000), scaled_vid)
                with open(scaled_fn, 'wb+') as f:
                    f.write(state.data)
                probe = ffmpeg.probe(scaled_fn)
                video_stream = next((stream for stream in probe['streams'] if stream['codec_type'] == 'video'), None)
                print(video_stream)
            else:
                print('Non-existing')
            print('####################################')