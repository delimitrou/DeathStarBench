import os
import sys
import json
import base64
import logging
import time
import queue
from pathlib import Path
# worker pool
from multiprocessing import Queue, Process
# dapr
from dapr.clients import DaprClient
from dapr.ext.grpc import App, InvokeMethodRequest, InvokeMethodResponse
from dapr.clients.grpc._state import StateOptions, Consistency, Concurrency
# import grpc
# ffmpeg
import ffmpeg
# prometheus
import prometheus_client
from prometheus_client import multiprocess, CollectorRegistry
import warnings
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
metaTopic       = os.getenv('META_TOPIC', 'meta')
sceneTopic      = os.getenv('SCENE_TOPIC', 'scene')
# state store
videoStore      = os.getenv('VIDEO_STORE', 'vpipe-video-store')
# worker
numWorkers      = int(os.getenv('WORKERS', '10'))
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
    'video_meta_prio_1_total', 
    'Number of priority 1 (high) meta requests')
# highPriDurTotal = prometheus_client.Counter(
#     'video_meta_prio_1_duration_total', 
#     'Total video duration of priority 1 (high) meta requests')
lowPriReqTotal = prometheus_client.Counter(
    'video_meta_prio_2_total', 
    'Number of priority 2 (low) meta requests')
# lowPriDurTotal = prometheus_client.Counter(
#     'video_meta_prio_2_duration_total', 
#     'Total video duration of priority 2 (low) meta requests')
# todo: the latency range needs refined
highPrioLat = prometheus_client.Histogram(
    'video_meta_prio_1_lat_hist',
    'Latency (ms) histogram of priority 1 (high) video-meta requests',
    buckets=pyutil.latBucketsLong()
)
lowPrioLat = prometheus_client.Histogram(
    'video_meta_prio_2_lat_hist',
    'Latency (ms) histogram of priority 2 (low) video-meta requests',
    buckets=pyutil.latBucketsLong()
)
videoStoreLat = prometheus_client.Histogram(
    'video_store_meta_read_lat_hist',
    'Latency (ms) histogram of reading video-store (kvs/db) in video-meta',
    buckets=pyutil.latBuckets()
)

MAX_PAYLOAD = 64 * 1024 * 1024 # 64MB

# meta extraction
def extractMeta(data, video_dir: Path):
    global MAX_PAYLOAD
    # dt = datetime.now(timezone.utc)
    video_id = data['video_id']
    req_id = data['req_id']
    send_unix_ms = data['send_unix_ms']
    is_high_prio = data['priority'] == 1
    # 1st stage, so client timestamp equal to send timestamp
    client_unix_ms = send_unix_ms
    # prom metrics
    epoch = time.time()*1000
    serv_lat = epoch - send_unix_ms 
    tempf = video_dir / video_id
    with DaprClient(max_grpc_message_length=MAX_PAYLOAD) as d:
        try:
            # logging.info('%s width=%d' %(data['data_id'], data['width']))
            video_b64 = d.get_state(
                store_name=videoStore, 
                key=video_id).data
            # update prom metrics
            cur_unix_ms = time.time()*1000
            store_lat = cur_unix_ms - epoch
            videoStoreLat.observe(store_lat)
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
        # logging.info('At %d metaWorker receives work: %s' %(
        #     int(time.time() * 1000), str(tempf)))
        probe = None
        try:
            probe = ffmpeg.probe(str(tempf))
        except ffmpeg.Error as e:
            logging.error('ffprobe stdout: %s, stderr: %s' %(e.stdout, e.stderr))
            return False
            # raise RuntimeError('ffprobe stdout: %s, stderr: %s' %(e.stdout, e.stderr))
        # remove temp files
        if os.path.exists(str(tempf)):
            os.remove(str(tempf))
        # get metadata
        duration = float(probe['format']['duration'])
        format = pyutil.pickFormat(probe['format']['format_name'])
        width = None
        video_stream = next((stream for stream in probe['streams'] if stream['codec_type'] == 'video'), None)
        if video_stream != None:
            width = video_stream['width']
        # update request & send to sceneShot service
        data['meta'] = {
            'duration': duration,
            'format': format,
            'width': width,
        }
        data['send_unix_ms'] = int(time.time() * 1000)
        data['client_unix_ms'] = int(client_unix_ms)
        # # todo: for debug only, comment off later
        # data['trace'] = {}
        # data['trace']['meta_serv_lat'] = time.time() * 1000 - epoch
        # data['trace']['meta_store_lat'] = store_lat
        
        resp = d.publish_event(
            pubsub_name=videoPipePubsub,
            topic_name=sceneTopic,
            data=json.dumps(data),
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
        logging.debug('Processed req_id=%s, video_id=%s, priority=%d, serv_lat=%d, store_lat=%d' %(
            req_id, video_id, data['priority'], serv_lat, store_lat,
        ))
        return True

# worker function
def metaWorker(
    worker_id: int,
    high_prio_queue: Queue,
    low_prio_queue: Queue,
    max_timeout: float = maxTimeout,
    idle_time: float = idleTime,
    interval: float=60,
    ):
    video_dir = Path('/tmp') / str(worker_id) / 'video'
    os.makedirs(str(video_dir), exist_ok=True)
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
                extractMeta(
                    data=req,
                    video_dir=video_dir)
                low_prio_ctr += 1
            except queue.Empty:
                high_prio_empty = True
                break
        # process low prioty requests when no high priority requests exist
        try:
            req = low_prio_queue.get(block=False)
            extractMeta(
                data=req,
                video_dir=video_dir)
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
@app.subscribe(pubsub_name=videoPipePubsub, topic=metaTopic)
def videoMeta(event) -> None:
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
        worker_p = Process(target=metaWorker, kwargs={
            'worker_id': i,
            'high_prio_queue': highPrioQueue,
            'low_prio_queue': lowPrioQueue,
        })
        worker_p.start()  # Launch reader_p() as another proc
        workerPool.append(worker_p)
    # start the service
    app.run(serviceAddress)