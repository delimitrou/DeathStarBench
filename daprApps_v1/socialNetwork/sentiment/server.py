import os
import sys
import json
import logging
import time
import numpy as np
# dapr
from dapr.clients import DaprClient
from dapr.ext.grpc import App
# ml
from transformers import pipeline
# prometheus
import prometheus_client
# util
from pathlib import Path
util_path = Path(__file__).parent.resolve() / '..' / 'pyutil'
sys.path.append(str(util_path))
import pyutil
# warnings
import warnings
warnings.filterwarnings("ignore")
# global variables
serviceAddress = int(os.getenv('ADDRESS', '5005'))
promAddress = int(os.getenv('PROM_ADDRESS', '8084'))
pubsubName = os.getenv('PUBSUB_NAME', 'sentiment-pubsub')
topicName = os.getenv('TOPIC_NAME', 'sentiment')
logging.basicConfig(level=logging.INFO)

# ml model
s0 = time.time()
sentiment = pipeline('sentiment-analysis')
logging.info(sentiment('Hello World')) # warm up the model
logging.info('Model download time: %.3fs' %(time.time() - s0))

# prometheus metrics
promReq = prometheus_client.Counter(
    'sentiment_processed', 
    'Number of sentiment requests processed',
)
# todo: the latency range needs refined
promLat = prometheus_client.Histogram(
    'sentiment_lat_hist',
    'Total latency of sentiment request (ms)',
    buckets=pyutil.latBucketsMl(),
)
promLatImg = prometheus_client.Histogram(
    'sentiment_img_lat_hist',
    'Total latency of sentiment (post w. image) request (ms)',
    buckets=pyutil.latBucketsMl(),
)
e2eSentiLat = prometheus_client.Histogram(
    'e2e_sentiment_lat_hist',
    'End-to-end latency of sentiment request (ms)',
    buckets=pyutil.latBucketsMl(),
)
e2eSentiImgLat = prometheus_client.Histogram(
    'e2e_sentiment_img_lat_hist',
    'End-to-end latency of sentiment request (ms), post w. image',
    buckets=pyutil.latBucketsMl(),
)

def newMetaReq(post_id: str, sentiment: list):
    jssent = json.dumps(sentiment)
    return {
        'post_id': post_id,
        'sentiment': jssent,
        'objects': None,
        'send_unix_ms': int(time.time() * 1000)
    }

# server 
app = App()
# handlers
@app.subscribe(pubsub_name=pubsubName, topic=topicName)
def sentimentAnalysis(event) -> None:
    global sentiment
    global promReq
    global promLat

    data = json.loads(event.Data())
    promReq.inc()
    # destination to publish processed results
    post_id = data['post_id']   # id of the entity that owns the images
    send_unix_ms = float(data['send_unix_ms'])
    image_included = bool(data['image_included'])
    client_unix_ms = float(data['client_unix_ms'])
    # inference
    pred = sentiment(data['text'])
    # update prom metrics
    cur_unix_ms = time.time() * 1000
    redeliver = False
    # redeliver not considered as queueing time
    if cur_unix_ms - send_unix_ms >= pyutil.redeliverInterval():
        redeliver = True
    if not redeliver:
        if image_included:
            promLatImg.observe(cur_unix_ms - send_unix_ms)
            e2eSentiImgLat.observe(cur_unix_ms - client_unix_ms)
        else:
            promLat.observe(cur_unix_ms - send_unix_ms)
            e2eSentiLat.observe(cur_unix_ms - client_unix_ms)
    logging.info('recv_unix_ms: %.1f, compl_unix_ms: %.1f, dur_ms=%.1f, sentiment=%s' %(
        send_unix_ms, cur_unix_ms, cur_unix_ms - send_unix_ms, str(pred),
    ))
    # wrap predictions into req
    meta_req = newMetaReq(post_id, pred)
    with DaprClient() as d:
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

# start prometheus
prometheus_client.start_http_server(promAddress)

# start the service
app.run(serviceAddress)