import os
import sys
import json
import logging
import time
import queue
import numpy as np
from pathlib import Path
from concurrent import futures
# worker pool
from multiprocessing import Queue, Process
from threading import Lock
# dapr
from dapr.clients import DaprClient
from dapr.ext.grpc import App, InvokeMethodRequest, InvokeMethodResponse
from dapr.clients.grpc._state import StateOptions, Consistency, Concurrency, StateItem
# import grpc
# opencv
import cv2
cascPath = cv2.data.haarcascades + 'haarcascade_frontalface_default.xml'
faceCascade = cv2.CascadeClassifier(cascPath)
# prometheus
import prometheus_client
from prometheus_client import multiprocess, CollectorRegistry
import warnings
from typing import Dict
# util
from pathlib import Path
util_path = Path(__file__).parent.resolve() / '..' / 'pyutil'
sys.path.append(str(util_path))
import pyutil

warnings.filterwarnings("ignore")
# global variables
serviceAddress  = int(os.getenv('ADDRESS', '5005'))
promAddress     = int(os.getenv('PROM_ADDRESS', '8084'))
# pubsub
videoPipePubsub = os.getenv('VIDEO_PIPE_PUBSUB', 'vpipe-events')
faceTopic       = os.getenv('FACE_TOPIC', 'face')
imageStore      = os.getenv('IMAGE_STORE', 'vpipe-image-store')
numWorkers      = int(os.getenv('WORKERS', '10'))
logLevel        = os.getenv('LOG_LEVEL', 'info')
maxTimeout      = float(os.getenv('MAX_TIMEOUT', '0.1'))
idleTime   = float(os.getenv('IDLE_TIME', '0.005'))

# multi-process prometheus dir (must exist and be empty)
# todo: create this dir in dockerfile and set the env in deploy yaml
prom_multiproc_dir = os.getenv('PROMETHEUS_MULTIPROC_DIR', '/tmp/prom_multiproc')
assert os.path.isdir(prom_multiproc_dir)
assert len(os.listdir(prom_multiproc_dir)) == 0
if logLevel == 'debug':
    logging.basicConfig(level=logging.DEBUG)
else:
    logging.basicConfig(level=logging.INFO)

# prometheus metrics
highPriReqTotal = prometheus_client.Counter(
    'face_detect_prio_1_total', 
    'Number of priority 1 (high) meta requests')
lowPriReqTotal = prometheus_client.Counter(
    'face_detect_prio_2_total', 
    'Number of priority 2 (low) meta requests')
# todo: the latency range needs refined
highPrioLat = prometheus_client.Histogram(
    'face_detect_prio_1_lat_hist',
    'Latency (ms) histogram of priority 1 (high) face-detect requests',
    buckets=pyutil.latBucketsLong()
)
lowPrioLat = prometheus_client.Histogram(
    'face_detect_prio_2_lat_hist',
    'Latency (ms) histogram of priority 2 (low) face-detect requests',
    buckets=pyutil.latBucketsLong()
)
# end-to-end latency
e2eHighPrioLat = prometheus_client.Histogram(
    'e2e_vpipe_prio_1_lat_hist',
    'End-to-end latency (ms) histogram of priority 1 (high) requests',
    buckets=pyutil.latBucketsLong()
)
e2eLowPrioLat = prometheus_client.Histogram(
    'e2e_vpipe_prio_2_lat_hist',
    'End-to-end latency (ms) histogram of priority 2 (low) requests',
    buckets=pyutil.latBucketsLong()
)
imageStoreReadLat = prometheus_client.Histogram(
    'image_store_face_read_lat_hist',
    'Latency (ms) histogram of reading video-store (kvs/db) in face-detect',
    buckets=pyutil.latBuckets()
)
imageStoreUpdateLat = prometheus_client.Histogram(
    'image_store_face_update_lat_hist',
    'Latency (ms) histogram of updating image-store (kvs/db) in face-detect',
    buckets=pyutil.latBuckets()
)

MAX_PAYLOAD = 64 * 1024 * 1024 # 64MB
# folders to hold videos
imageDir = Path('/tmp') / 'image'
os.makedirs(str(imageDir), exist_ok=True)

# meta extraction
def faceDetect(data):
    # dt = datetime.now(timezone.utc)
    image_id = data['image_id']
    send_unix_ms = data['send_unix_ms']
    client_unix_ms = data['client_unix_ms']
    is_high_prio = data['priority'] == 1
    # prom metrics
    epoch = time.time()*1000
    serv_lat = epoch - send_unix_ms
    read_store_lat = 0
    image = None
    with DaprClient(max_grpc_message_length=MAX_PAYLOAD) as d:
        max_trial = 3
        data_fetched = False
        trials = 0
        while True:
            try:
                # logging.info('%s width=%d' %(data['data_id'], data['width']))
                ts_start = time.time()*1000
                image_bytes = d.get_state(
                    store_name=imageStore, 
                    key=image_id).data
                ts_end = time.time()*1000
                read_store_lat += ts_end - ts_start
                if len(image_bytes) > 0:
                    # https://stackoverflow.com/questions/33548956/detect-avoid-premature-end-of-jpeg-in-cv2-python              
                    check_chars = image_bytes[-2:]
                    if check_chars != b'\xff\xd9':
                        serv_lat += time.time()*1000 - ts_end
                        raise ValueError('Incomplete image data')
                    else:
                        image_np = np.frombuffer(image_bytes, np.uint8)
                        image = cv2.imdecode(image_np, cv2.IMREAD_COLOR) 
                        serv_lat += time.time()*1000 - ts_end
                        data_fetched = True
            except cv2.error as e:
                logging.error('Failed to read %s from %s: %s' %(
                    image_id, imageStore, str(e)
                ))
            except Exception as e:
                logging.error('Failed to read %s from %s: %s' %(
                    image_id, imageStore, str(e)
                ))
            trials += 1
            if data_fetched or trials >= max_trial:
                break
            else:
                time.sleep(0.01)
        if not data_fetched or image is None:
            logging.error('Key: %s is not available in %s' %(
                image_id, imageStore
            ))
            return False
        # parse video data
        epoch = time.time()*1000
        try:
            gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
            # Detect faces in the image
            faces = faceCascade.detectMultiScale(
                gray,
                scaleFactor=1.1,
                minNeighbors=5,
                minSize=(30, 30),
                flags = cv2.CASCADE_SCALE_IMAGE
            )
            # Draw a rectangle around the faces
            for (x, y, w, h) in faces:
                cv2.rectangle(image, (x, y), (x+w, y+h), (0, 255, 0), 2)
            image_bytes = cv2.imencode('.jpg', image)[1].tobytes()
        except cv2.error as e:
            logging.error('Failed detect face in image %s for cv2 error: %s' %(
                image_id, str(e)
            ))
            return False
        except Exception as e:
            logging.error('Failed to detect face in image %s: %s' %(
                image_id, str(e)
            ))
        # update prom metrics
        imageStoreReadLat.observe(read_store_lat)
        # latency metrics
        cur_unix_ms = time.time()*1000
        serv_lat += cur_unix_ms - epoch
        epoch = cur_unix_ms
        d.save_state(
            store_name=imageStore, 
            key=image_id, 
            value=image_bytes,
            options=StateOptions(
                consistency=Consistency.strong,
                concurrency=Concurrency.last_write,
            ),
        )
        # latency metrics
        cur_unix_ms = time.time()*1000
        img_store_lat = cur_unix_ms - epoch
        epoch = cur_unix_ms
        imageStoreUpdateLat.observe(img_store_lat)
        cur_unix_ms = time.time()*1000
        serv_lat += cur_unix_ms - epoch
        if is_high_prio:
            highPrioLat.observe(serv_lat)
            e2eHighPrioLat.observe(cur_unix_ms - client_unix_ms)
            # if cur_unix_ms - client_unix_ms >= 20000:
            #     logging.info('high_prio_req e2e_lat = %dms, meta_serv=%d, meta_store=%d, scene_serv=%d, scene_store=%d, face_serv=%d, face_store=%d' %(
            #         cur_unix_ms - client_unix_ms, 
            #         data['trace']['meta_serv_lat'],
            #         data['trace']['meta_store_lat'],
            #         data['trace']['scene_serv_lat'],
            #         data['trace']['scene_store_lat'],
            #         serv_lat,
            #         read_store_lat,
            #     ))
        else:
            lowPrioLat.observe(serv_lat) 
            e2eLowPrioLat.observe(cur_unix_ms - client_unix_ms)
        
        logging.debug('Processed req_id=%s, image_id=%s, priority=%d, serv=%d, read_store=%d, write_store=%d, e2e=%d' %(
            data['req_id'], 
            image_id, 
            data['priority'], 
            serv_lat, 
            read_store_lat, 
            img_store_lat,
            cur_unix_ms - client_unix_ms,
        ))
        return True

# worker function
def faceDetectWorker(
    high_prio_queue: Queue,
    low_prio_queue: Queue,
    max_timeout: float = maxTimeout,
    idle_time: float = idleTime,
    interval: float=60,
    ):
    low_prio_ctr = 0
    low_prio_rps = 0
    stats_ts = time.time()
    ts_empty = True
    timeout = 0.05
    while True:
        # inifinite loop processing both high & low priority requests
        high_prio_empty = False
        low_prio_empty = False
        # process high priority requests as long as there is any
        while True:
            epoch = time.time()
            # todo: compute propoer timeout
            if epoch - stats_ts >= interval:
                rps = low_prio_ctr / (epoch - stats_ts)
                if ts_empty:
                    low_prio_rps = rps
                    ts_empty = False
                else:
                    low_prio_rps = rps * 0.4 + low_prio_rps * 0.6
                if rps > 0:
                    timeout = min(idle_time / low_prio_rps, max_timeout)
                else:
                    timeout = max_timeout
                logging.debug('low_prio_rps = %.3f, timeout set to %.3fs' %(
                    low_prio_rps, timeout))
                low_prio_ctr = 0
                stats_ts = epoch
            # actual work
            try:
                # req = high_prio_queue.get(block=False)
                req = high_prio_queue.get(block=True, timeout=timeout)
                faceDetect(data=req)
            except queue.Empty:
                high_prio_empty = True
                break
        # process low prioty requests when no high priority requests exist
        try:
            req = low_prio_queue.get(block=False)
            faceDetect(data=req)
            low_prio_ctr += 1
        except queue.Empty:
            low_prio_empty = True
        # wait if no req is pending
        if high_prio_empty and low_prio_empty:
            time.sleep(0.01)

workerPool = [] # list of worker processes
highPrioQueue = Queue()
lowPrioQueue = Queue()

app = App()
# upload a new video
@app.subscribe(pubsub_name=videoPipePubsub, topic=faceTopic)
def videoScene(event) -> None:
    # request queues
    global highPrioQueue
    global lowPrioQueue
    data = json.loads(event.Data())
    if data['priority'] == 1:
        # high priority requests
        highPriReqTotal.inc()
        highPrioQueue.put(data)
    elif data['priority'] == 2:
        lowPriReqTotal.inc()
        # low priority requests
        lowPrioQueue.put(data)

if __name__ == '__main__':
    # create multiprocess registry and start prometheus service
    registry = CollectorRegistry()
    multiprocess.MultiProcessCollector(registry)
    prometheus_client.start_http_server(promAddress, registry=registry)
    # worker pool
    for i in range(0, numWorkers):
        worker_p = Process(target=faceDetectWorker, kwargs={
            'high_prio_queue': highPrioQueue,
            'low_prio_queue': lowPrioQueue,
        })
        worker_p.start()  # Launch reader_p() as another proc
        workerPool.append(worker_p)
    # start the service
    app.run(serviceAddress)