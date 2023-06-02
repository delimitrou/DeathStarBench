import os
import sys
import json
import logging
import time
from pathlib import Path
# dapr
from dapr.clients import DaprClient
from dapr.ext.grpc import App
# ml
from transformers import pipeline
# image
from PIL import Image
import io
# prometheus
import prometheus_client
import warnings
from typing import Dict
# util
from pathlib import Path
util_path = Path(__file__).parent.resolve() / '..' / 'pyutil'
sys.path.append(str(util_path))
import pyutil

warnings.filterwarnings("ignore")
# global variables
serviceAddress = int(os.getenv('ADDRESS', '5005'))
promAddress = int(os.getenv('PROM_ADDRESS', '8084'))
pubsubName = os.getenv('PUBSUB_NAME', 'object-detect-pubsub')
topicName = os.getenv('TOPIC_NAME', 'object-detect')
imageStore = os.getenv('IMAGE_STORE', 'image-store')
logging.basicConfig(level=logging.INFO)

# ml model
s0 = time.time()
objectDetector = pipeline('object-detection')
# objectDetector = pipeline(model='mishig/tiny-detr-mobilenetsv3')
# todo: do we need to warm up here since the model is already pulled
# logging.info(objectDetector('Hello World')) # warm up the model
logging.info('Model download time: %.3fs' %(time.time() - s0))

# prometheus metrics
promReq = prometheus_client.Counter(
    'object_detect_processed', 
    'Number of object detections requests processed')
# todo: the latency range needs refined
servLat = prometheus_client.Histogram(
    'object_detect_serv_lat_hist',
    'Latency (ms) histogram of object-detect requests, excluding time waiting for kvs/db',
    buckets=pyutil.latBucketsLongMl()
)
storeLat = prometheus_client.Histogram(
    'img_store_read_lat_hist',
    'Latency (ms) histogram of reading img store (kvs/db).',
    buckets=pyutil.latBuckets()
)
e2eObjDetLat = prometheus_client.Histogram(
    'e2e_object_detect_lat_hist',
    'End-to-end latency (ms) histogram of object detection.',
    buckets=pyutil.latBucketsLongMl()
)

def newMetaReq(post_id: str, objects: dict):
    jsobj = {}
    for o in objects:
        jsobj[o] = json.dumps(objects[o])
    return {
        'post_id': post_id,
        'sentiment': '',
        'objects': jsobj,
        'send_unix_ms': int(time.time() * 1000)
    }

# server 
app = App()
# handlers
# @app.method(name='object_detect')
@app.subscribe(pubsub_name=pubsubName, topic=topicName)
def objectDetect(event) -> None:
    global objectDetector
    global promReq
    global servLat

    data = json.loads(event.Data())
    post_id = data['post_id']
    send_unix_ms = float(data['send_unix_ms'])
    client_unix_ms = float(data['client_unix_ms'])
    # latency metrics
    epoch = time.time()*1000
    serv_lat = epoch - send_unix_ms
    redeliver = False
    # redeliver not considered as queueing time
    if serv_lat >= pyutil.redeliverInterval():
        redeliver = True
    if len(data['images']) > 0:
        promReq.inc()
        pil_images = []
        with DaprClient() as d:
            try:
                logging.info(data['images'])
                items = d.get_bulk_state(
                    store_name=imageStore, 
                    keys=data['images'], 
                ).items
                # update prom metrics
                cur_unix_ms = time.time()*1000
                storeLat.observe(cur_unix_ms - epoch)
                epoch = cur_unix_ms
                for i in items:
                    # etag = i.etag
                    # logging.info(type(i.data))
                    # logging.info(len(i.data))
                    # logging.info(sys.getsizeof(i.data))
                    # logging.info('len=%d, size=%d' %(len(i.data), sys.getsizeof(i.data)))
                    img = Image.open(io.BytesIO(i.data))
                    pil_images.append(img)
            except Exception as e:
                logging.error('Failed to read from %s: %s' %(
                    imageStore, str(e)
                ))
                return
            # inference
            pred = objectDetector(pil_images)
            # update prom metrics
            if not redeliver:
                cur_unix_ms = time.time() * 1000
                serv_lat += cur_unix_ms - epoch
                servLat.observe(serv_lat)
                e2eObjDetLat.observe(cur_unix_ms - client_unix_ms)
            # wrap predictions into events
            objects = {}
            # logging.info('recv_unix_ms: %.1f, compl_unix_ms: %.1f, dur_ms=%.1f' %(
            #     send_unix_ms, cur_unix_ms, cur_unix_ms - send_unix_ms,
            # ))
            for img_id, p in zip(data['images'], pred):
                # todo: for testing only, comment off later
                logging.info("%s: %s" %(img_id, str(p)))
                """
                each prediction should be a list of dicts,
                and each dict has the following keys: 
                    label (str) — The class label identified by the model.
                    score (float) — The score attributed by the model for that label.
                    box (List[Dict[str, int]]) — The bounding box of detected object in image original size.
                """
                objects[img_id] = p
            meta_req = newMetaReq(post_id, objects)
            try:    
                resp = d.invoke_method(
                    'dapr-post',
                    'meta',
                    data=json.dumps(meta_req),
                )
                # todo: comment off later
                logging.info(resp.headers)
            except Exception as e:
                logging.error('Failed to invoke dapr-post:meta %s' %str(e))
                
    else:
        logging.warning('Empty event with no images')

# start prometheus
prometheus_client.start_http_server(promAddress)

# start the service
app.run(serviceAddress)