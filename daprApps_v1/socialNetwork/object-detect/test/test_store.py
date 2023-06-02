import sys
import os
import time
import threading
import random
import json
import logging
from pathlib import Path
import argparse
from dapr.clients import DaprClient
from dapr.clients.grpc._state import StateItem, StateOptions, Consistency, Concurrency
from PIL import Image
import io

logging.basicConfig(level=logging.INFO)
parser = argparse.ArgumentParser()
parser.add_argument('--s', dest='save_image', action='store_true')
args = parser.parse_args()
save_image = args.save_image

pubsub_name = 'object-detect-pubsub'
topic_name = 'object-detect'
img_store = os.getenv('IMAGE_STORE', 'image-store-test')
post_store = os.getenv('POST_STORE', 'post-store-test')

# save states
images = [
    'panda2.jpg',  
    'panda.jpeg',  
    'shiba2.jpg',  
    'shiba.jpg',
    ]

local_data = {}

def get_images(image_ids):
    with DaprClient() as d:
        items = d.get_bulk_state(
            store_name=img_store, 
            keys=image_ids, 
        ).items
        pil_images = []
        for idx, i in enumerate(items):
            img_name = image_ids[idx]
            etag = i.etag
            # logging.info(image_ids[idx])
            # logging.info(type(i.data))
            # logging.info(len(i.data), len(local_data[img_name]))
            # logging.info(sys.getsizeof(i.data), sys.getsizeof(local_data[img_name]))
            logging.info('%s, len=%d, local_len=%d; size=%d, local_size=%d' %(
                image_ids[idx], len(i.data), len(local_data[img_name]), 
                sys.getsizeof(img_data), sys.getsizeof(local_data[img_name])
                ))
            # img = Image.open(io.BytesIO(i.data))
            # pil_images.append(img)


with DaprClient() as d:
    for img in images:
        with open('images/' + img, 'rb') as f:
            img_data = f.read()
            local_data[img] = img_data
            logging.info('%s, len=%d, size=%d' %(img, len(img_data), sys.getsizeof(img_data)))
            pil_img = Image.open(io.BytesIO(img_data))
            print(pil_img)
            if save_image:
                resp = d.save_state(
                    store_name=img_store, 
                    key=img, 
                    value=img_data,
                    options=StateOptions(consistency=Consistency.strong),
                    )
                print(resp.headers)
                print('%s saved' %img)
    
    # # test save_bulk_state
    # if save_image:
    #     all_states = []
    #     for img in local_data:
    #         all_states.append(StateItem(key=img, value=local_data[img]))
    #     resp = d.save_bulk_state(
    #         store_name=img_store, 
    #         states=all_states,
    #         )
    #     print(resp.headers)
    #     print('all images saved')

get_images(images)
time.sleep(5)
get_images(images)
time.sleep(10)
get_images(images)