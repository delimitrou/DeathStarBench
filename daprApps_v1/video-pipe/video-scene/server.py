from genericpath import isdir
import os
import sys
import json
import base64
import logging
import time
import copy
import queue
from pathlib import Path
from concurrent import futures
import subprocess
# import subprocess
# import shutil
# worker pool
from multiprocessing import Queue, Process
# dapr
from dapr.clients import DaprClient
from dapr.ext.grpc import App, InvokeMethodRequest, InvokeMethodResponse
from dapr.clients.grpc._state import StateOptions, Consistency, Concurrency, StateItem
# import grpc
# ffmpeg
import ffmpeg
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
serviceAddress = int(os.getenv('ADDRESS', '5005'))
promAddress    = int(os.getenv('PROM_ADDRESS', '8084'))
# pubsub
videoPipePubsub = os.getenv('VIDEO_PIPE_PUBSUB', 'vpipe-events')
sceneTopic      = os.getenv('SCENE_TOPIC', 'scene')
faceTopic     = os.getenv('FACE_TOPIC', 'face')
# state store
videoStore     = os.getenv('VIDEO_STORE', 'vpipe-video-store')
imageStore     = os.getenv('IMAGE_STORE', 'vpipe-image-store')
# worker
numWorkers     = int(os.getenv('WORKERS', '10'))
# scene configuration
sceneInterval  = int(os.getenv('SCENE_INTERVAL', '5'))
maxScenes      = int(os.getenv('MAX_SCENES', '4'))
logLevel        = os.getenv('LOG_LEVEL', 'info')
maxTimeout      = float(os.getenv('MAX_TIMEOUT', '0.1'))
idleTime   = float(os.getenv('IDLE_TIME', '0.001'))

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
    'video_scene_prio_1_total', 
    'Number of priority 1 (high) meta requests')
highPriDurTotal = prometheus_client.Counter(
    'video_scene_prio_1_duration_total', 
    'Total video duration of priority 1 (high) scene-extraction requests')
lowPriReqTotal = prometheus_client.Counter(
    'video_scene_prio_2_total', 
    'Number of priority 2 (low) meta requests')
lowPriDurTotal = prometheus_client.Counter(
    'video_meta_prio_2_duration_total', 
    'Total video duration of priority 2 (low) scene-extraction requests')
# todo: the latency range needs refined
highPrioLat = prometheus_client.Histogram(
    'video_scene_prio_1_lat_hist',
    'Latency (ms) histogram of priority 1 (high) video-scene requests',
    buckets=pyutil.latBucketsLong()
)
lowPrioLat = prometheus_client.Histogram(
    'video_scene_prio_2_lat_hist',
    'Latency (ms) histogram of priority 2 (low) video-scene requests',
    buckets=pyutil.latBucketsLong()
)
videoStoreLat = prometheus_client.Histogram(
    'video_store_scene_read_lat_hist',
    'Latency (ms) histogram of reading video-store (kvs/db) in video-scene',
    buckets=pyutil.latBuckets()
)
imageStoreLat = prometheus_client.Histogram(
    'image_store_scene_update_lat_hist',
    'Latency (ms) histogram of updating image-store (kvs/db) in video-scene',
    buckets=pyutil.latBuckets()
)

MAX_PAYLOAD = 64 * 1024 * 1024 # 64MB

# meta extraction
def extractScene(data, video_dir: Path, image_dir: Path):
    # dt = datetime.now(timezone.utc)
    video_id = data['video_id']
    req_id = data['req_id']
    send_unix_ms = data['send_unix_ms']
    is_high_prio = data['priority'] == 1
    duration = data['meta']['duration']
    # update prom workload metric
    if is_high_prio:
        highPriDurTotal.inc(duration)
    else:
        lowPriDurTotal.inc(duration)
    # prom metrics
    epoch = time.time()*1000
    serv_lat = epoch - send_unix_ms 
    tmp_video_id = '%s-%s' %(req_id, video_id)
    tempf = video_dir / tmp_video_id
    with DaprClient(max_grpc_message_length=MAX_PAYLOAD) as d:
        try:
            # logging.info('%s width=%d' %(data['data_id'], data['width']))
            video_b64 = d.get_state(
                store_name=videoStore, 
                key=video_id).data
            # update prom metrics
            cur_unix_ms = time.time()*1000
            video_store_lat = cur_unix_ms - epoch
            videoStoreLat.observe(video_store_lat)
            # parse video data
            epoch = cur_unix_ms
            video_bytes = base64.b64decode(video_b64)
            with open(str(tempf), 'wb+') as f:
                f.write(video_bytes)
        except Exception as e:
            logging.error('Failed to read %s from %s: %s' %(
                video_id, videoStore, str(e)
            ))
            return False
        
        # get metadata of the video
        # logging.info('At %d sceneWorker receives work: %s' %(
        #     int(time.time() * 1000), str(tempf)))
        temp_img_dir = image_dir / tmp_video_id
        if not os.path.isdir(temp_img_dir):
            os.makedirs(temp_img_dir)
        try:
            # take snapshot every few seconds
            rate = max(1/sceneInterval, maxScenes/duration)
            (
                ffmpeg
                .input(str(tempf))
                .output(str(temp_img_dir) + '/' + str(req_id) + '_%02d.jpg', 
                    r=rate, format='image2', vcodec='mjpeg')
                .overwrite_output()
                # .run_async(pipe_stdout=True, pipe_stderr=True)
                .run(capture_stdout=True, capture_stderr=True)
            )   
        except ffmpeg.Error as e:
            out = e.stdout.decode()
            err = e.stderr.decode()
            logging.error('FFmpeg (req_id: %s) std_err: %s, std_out: %s' %(
                req_id, err, out))
            if os.path.exists(str(tempf)):
                os.remove(str(tempf))
            return False
            # raise RuntimeError('ffprobe stdout: %s, stderr: %s' %(e.stdout, e.stderr))
        # remove video data
        if os.path.exists(str(tempf)):
            os.remove(str(tempf))
        state_items = []
        image_ids = []
        # exclude the image at the beginning and the image at the end
        num_images = 0
        ignored_images = []
        max_img_id = ''
        for img in os.listdir(str(temp_img_dir)):
            this_img_id = img.split('_')[-1].replace('.jpg', '')
            if max_img_id == '' or int(this_img_id.lstrip('0')) > int(max_img_id.lstrip('0')):
                max_img_id = this_img_id
            num_images += 1
        if num_images > 2:
            ignored_images = [str(req_id) + '_01.jpg', str(req_id) + '_%s.jpg' %max_img_id]
            logging.debug('ignored images: %s' %(','.join(ignored_images)))
        for img in os.listdir(str(temp_img_dir)):
            if img in ignored_images:
                logging.debug('images:%s ignored' %img)
                continue
            img_id = img.split('/')[-1]
            if os.path.isfile(str(temp_img_dir / img)):
                image_ids.append(img_id)
                with open(str(temp_img_dir / img), 'rb') as f:
                    img_bytes = f.read()
                    # logging.info('Image: %s, length_bytes=%d' %(
                    #     img_id, len(img_bytes)))
                    state_items.append(
                        StateItem(
                            key=img_id,
                            value=img_bytes,
                            # options=StateOptions(
                            #     consistency = Consistency.strong,
                            #     concurrency = Concurrency.first_write,
                            # )
                        )
                    )
        # latency metrics
        cur_unix_ms = time.time()*1000
        serv_lat += cur_unix_ms - epoch
        epoch = cur_unix_ms
        # save state
        if len(state_items) > 0:
            d.save_bulk_state(
                store_name=imageStore, 
                states=state_items,
            )
        # latency metrics
        cur_unix_ms = time.time()*1000
        img_store_lat = cur_unix_ms - epoch
        imageStoreLat.observe(img_store_lat)
        epoch = cur_unix_ms
        # remove temp video directory
        subprocess.run('rm -rf %s' %(str(temp_img_dir)), shell=True)
        # shutil.rmtree(str(temp_img_dir))
        # todo: send one request per image
        if len(image_ids) > 0:
            for img_id in image_ids:
                req_data = copy.copy(data)
                req_data['send_unix_ms'] = int(time.time()*1000)
                req_data['image_id'] = img_id

                # # todo: debug, remove later
                # req_data['trace'] = data['trace']
                # req_data['trace']['scene_serv_lat'] = serv_lat + time.time() * 1000 - epoch
                # req_data['trace']['scene_store_lat'] = video_store_lat

                resp = d.publish_event(
                    pubsub_name=videoPipePubsub,
                    topic_name=faceTopic,
                    data=json.dumps(req_data),
                    data_content_type='application/json',
                )
        # update prom metrics
        cur_unix_ms = time.time() * 1000
        serv_lat += cur_unix_ms - epoch
        if is_high_prio:
            highPrioLat.observe(serv_lat)
        else:
            lowPrioLat.observe(serv_lat)
        # for debugging
        logging.debug('Processed req_id=%s, video_id=%s, images=%d, priority=%d, serv_lat=%d, video_store_lat=%d, img_store_lat=%d' %(
            req_id, video_id, len(image_ids), data['priority'], serv_lat, video_store_lat, img_store_lat,
        ))
        return True

# worker function
def sceneWorker(
    worker_id: int,
    high_prio_queue: Queue,
    low_prio_queue: Queue,
    max_timeout: float = maxTimeout,
    idle_time: float = idleTime,
    interval: float=60,
    ):
    video_dir = Path('/tmp') / str(worker_id) / 'video'
    image_dir = Path('/tmp') / str(worker_id) / 'image'
    os.makedirs(str(video_dir), exist_ok=True)
    os.makedirs(str(image_dir), exist_ok=True)
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
                extractScene(
                    data=req,
                    video_dir=video_dir,
                    image_dir=image_dir)
            except queue.Empty:
                high_prio_empty = True
                break
        # process low prioty requests when no high priority requests exist
        try:
            req = low_prio_queue.get(block=False)
            extractScene(data=req,
                video_dir=video_dir,
                image_dir=image_dir)
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
@app.subscribe(pubsub_name=videoPipePubsub, topic=sceneTopic)
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
        worker_p = Process(target=sceneWorker, kwargs={
            'worker_id': i,
            'high_prio_queue': highPrioQueue,
            'low_prio_queue': lowPrioQueue,
        })
        worker_p.start()  # Launch reader_p() as another proc
        workerPool.append(worker_p)
    # start the service
    app.run(serviceAddress)