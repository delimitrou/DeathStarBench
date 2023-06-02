import os
import sys
import json
import logging
import time
import random
from pathlib import Path
from concurrent import futures
from threading import Lock
# worker pool
from multiprocessing import Pool, get_context
# dapr
from dapr.clients import DaprClient
from dapr.ext.grpc import App
# ffmpeg
import ffmpeg
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
pubsubName = os.getenv('PUBSUB_NAME', 'video-pubsub')
topicName = os.getenv('TOPIC_NAME', 'thumbnail')
videoStore = os.getenv('VIDEO_STORE', 'video-store')
thumbnailStore = os.getenv('THUMBNAIL_STORE', 'thumbnail-store')
numWorkers = int(os.getenv('WORKERS', '10'))
logLevel    = os.getenv('LOG_LEVEL', 'info')
if logLevel == 'debug':
    logging.basicConfig(level=logging.DEBUG)
else:
    logging.basicConfig(level=logging.INFO)

# prometheus metrics
promReq = prometheus_client.Counter(
    'video_thumbnail_processed_total', 
    'Number of video-thumbnail requests processed')
# todo: the latency range needs refined
servLat = prometheus_client.Histogram(
    'video_thumbnail_serv_lat_hist',
    'Latency (ms) histogram of video-thumbnail requests (since upstream sends the req), excluding time waiting for kvs/db',
    buckets=pyutil.latBucketsFFmpegThumb()
)
storeLat = prometheus_client.Histogram(
    'video_thumbnail_store_lat_hist',
    'Latency (ms) histogram of reading & writing video store (kvs/db) for video_thumbnail requests',
    buckets=pyutil.latBuckets()
)
e2eVideoThumbnailLat = prometheus_client.Histogram(
    'e2e_video_thumbnail_lat_hist',
    'End-to-end latency (ms) histogram of video-thumbnail.',
    buckets=pyutil.latBucketsFFmpegThumb()
)

# folders to hold videos
dataDir = Path('/tmp') / 'video'
os.makedirs(str(dataDir), exist_ok=True)
# worker function
def videoProcessor(req):
    data_id = req['data_id']
    duration = req['duration']
    video_path = req['video_path']
    thumbnail_path = req['thumbnail_path']

    resp = {
        'succ': False,
        'error': '',
    }
    # generate thumbnail #
    ss = min(0.1, duration/10)
    try:
        (
            ffmpeg
            .input(str(video_path), ss=ss)
            .output(str(thumbnail_path), vframes=1, format='image2', vcodec='mjpeg')
            .overwrite_output()
            .run(quiet=True)
            # .run(capture_stdout=True)
        )     
        resp['succ'] = True  
    except ffmpeg.Error as e:
        out = e.stdout.decode()
        err = e.stderr.decode()
        resp['error'] = 'FFmpeg (data_id: %s) std_err: %s, std_out: %s' %(
            data_id, err, out)
    return resp
# worker pool
workerPool = None

# server 
MAX_PAYLOAD = 64 * 1024 * 1024 # 64MB
# MAX_PAYLOAD=-1
executor = futures.ThreadPoolExecutor(max_workers=20)
grpcOptions = [
    ('grpc.max_send_message_length', MAX_PAYLOAD),
    ('grpc.max_receive_message_length', MAX_PAYLOAD),
]
app = App(
    thread_pool=executor, 
    options=grpcOptions,
)

reqCtr = 0
reqCtrLock = Lock()

def getCtr():
    global reqCtr, reqCtrLock
    ctr = 0
    with reqCtrLock:
        ctr = reqCtr
        reqCtr += 1
    return ctr

# handlers
# @app.method(name='object_detect')
@app.subscribe(pubsub_name=pubsubName, topic=topicName)
def videoThumbnail(event) -> None:
    global promReq
    global servLat
    global workerPool
    # update req counter
    promReq.inc()
    data = json.loads(event.Data())
    video_id = data['video_id']
    data_id = data['data_id']
    duration = data['duration']
    send_unix_ms = float(data['send_unix_ms'])
    # client_unix_ms = float(data['client_unix_ms'])
    # latency metrics
    epoch = time.time()*1000
    serv_lat = epoch - send_unix_ms
    store_lat = 0
    # temp files saving video, add a random number to avoid conflict
    unique_id = getCtr()
    video_path = dataDir / ('%s-%d' %(data_id, unique_id))
    thumbnail_id = pyutil.thumbnailId(video_id)
    thumbnail_path = dataDir / ('%d-%s' %(unique_id, thumbnail_id))
    with DaprClient(max_grpc_message_length=MAX_PAYLOAD) as d:
        try:
            logging.debug('%s -> %s' %(data_id, thumbnail_id))
            logging.debug('%s -> %s' %(str(video_path), str(thumbnail_path)))
            video = d.get_state(
                store_name=videoStore, 
                key=data_id).data
            # update prom metrics
            cur_unix_ms = time.time()*1000
            store_lat += cur_unix_ms - epoch
            epoch = cur_unix_ms
            # return if video does not exist
            if len(video) == 0:
                logging.error('Cannot find video: %s in %s, or video is empty' %(
                    data_id, videoStore))
                # update prom metrics 
                promReq.inc()
                storeLat.observe(store_lat)
                servLat.observe(serv_lat)
                e2eVideoThumbnailLat.observe(cur_unix_ms - send_unix_ms)
                return
            # write video to a file
            with open(str(video_path), 'wb+') as f:
                f.write(video)
        except Exception as e:
            logging.error('Failed to read %s from %s: %s' %(
                data_id, videoStore, str(e)
            ))
            return
        # dispatch to worker pool
        work = {
            'data_id': data_id,
            'duration': duration,
            'video_path': video_path,
            'thumbnail_path': thumbnail_path,
        }
        fresult = workerPool.apply_async(videoProcessor, (work,))
        result = fresult.get()
        if not result['succ']:
            logging.error('FFmpeg error: %s' %result['error'])
        else:
            cur_unix_ms = time.time()*1000
            serv_lat += cur_unix_ms - epoch
            epoch = cur_unix_ms
            logging.debug('video_thumbnail serv dur_ms=%.1f' %(serv_lat))
            # save the scaled video
            with open(str(thumbnail_path), 'rb') as f:
                thumbnail = f.read()
                # logging.info('size of thumbnail=%d' %len(thumbnail))
                resp = d.save_state(
                    store_name=thumbnailStore, 
                    key=thumbnail_id, 
                    value=thumbnail
                )
                # logging.info(resp.headers)
            # update latency metric
            cur_unix_ms = time.time() * 1000
            store_lat += cur_unix_ms - epoch
            epoch = cur_unix_ms
        # remove temp files
        if os.path.exists(str(video_path)):
            os.remove(str(video_path))
        if os.path.exists(str(thumbnail_path)):
            os.remove(str(thumbnail_path))                        
        # update prom metrics 
        serv_lat += time.time() * 1000 - epoch
        storeLat.observe(store_lat)
        servLat.observe(serv_lat)
        logging.debug('e2e lat = %.1fms' %(cur_unix_ms - send_unix_ms))
        # logging.info('---------------------------------------')
        e2eVideoThumbnailLat.observe(cur_unix_ms - send_unix_ms)

if __name__ == '__main__':
    # worker pool
    workerPool = get_context("spawn").Pool(processes=numWorkers)
    # start prometheus
    prometheus_client.start_http_server(promAddress)
    # start the service
    app.run(serviceAddress)