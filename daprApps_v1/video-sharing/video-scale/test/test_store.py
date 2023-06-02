import sys
import time
import logging
from pathlib import Path
import argparse
from dapr.clients import DaprClient
from dapr.clients.grpc._state import StateItem, StateOptions, Consistency, Concurrency

util_path = Path(__file__).parent.resolve() / 'dapr_client'
sys.path.append(str(util_path))
from dapr_client import MyDaprClient

logging.basicConfig(level=logging.INFO)
parser = argparse.ArgumentParser()
parser.add_argument('--s', dest='save_video', action='store_true')
args = parser.parse_args()
save_video = args.save_video
video_store = 'video-store-test'

# save states
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
local_data = {}

MAX_MESSAGE_LENGTH=33554432
grpcOptions = [
    ('grpc.max_send_message_length', MAX_MESSAGE_LENGTH),
    ('grpc.max_receive_message_length', MAX_MESSAGE_LENGTH),
]

def get_videos(video_ids):
    global grpcOptions
    with MyDaprClient(options=grpcOptions) as d:
        items = d.get_bulk_state(
            store_name=video_store, 
            keys=video_ids, 
        ).items
        for idx, i in enumerate(items):
            video_name = video_ids[idx]
            etag = i.etag
            # logging.info(video_ids[idx])
            # logging.info(type(i.data))
            # logging.info(len(i.data), len(local_data[video_name]))
            # logging.info(sys.getsizeof(i.data), sys.getsizeof(local_data[video_name]))
            logging.info('%s, len=%d, local_len=%d; size=%d, local_size=%d' %(
                video_ids[idx], len(i.data), len(local_data[video_name]), 
                sys.getsizeof(video_data), sys.getsizeof(local_data[video_name])
                ))
            # video = Image.open(io.BytesIO(i.data))
            # pil_videos.append(video)


with MyDaprClient(options=grpcOptions) as d:
    for video in videos:
        with open('video/' + video, 'rb') as f:
            video_data = f.read()
            local_data[video] = video_data
            logging.info('%s, len=%d, size=%d' %(video, len(video_data), sys.getsizeof(video_data)))
            if save_video:
                resp = d.save_state(
                    store_name=video_store, 
                    key=video, 
                    value=video_data,
                    options=StateOptions(consistency=Consistency.strong),
                    )
                print(resp.headers)
                print('%s saved' %video)
    
    # # test save_bulk_state
    # if save_video:
    #     all_states = []
    #     for video in local_data:
    #         all_states.append(StateItem(key=video, value=local_data[video]))
    #     resp = d.save_bulk_state(
    #         store_name=video_store, 
    #         states=all_states,
    #         )
    #     print(resp.headers)
    #     print('all videos saved')

time.sleep(10)
get_videos(videos)