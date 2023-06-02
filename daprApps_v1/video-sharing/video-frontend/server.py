import os
import sys
import json
import base64
import logging
import time
from datetime import datetime, timezone
import re
from pathlib import Path
from concurrent import futures
from threading import Lock
# worker pool
from multiprocessing import Pool, get_context
# dapr
from dapr.clients import DaprClient
from dapr.ext.grpc import App, InvokeMethodRequest, InvokeMethodResponse
from dapr.clients.grpc._state import StateOptions, Consistency, Concurrency
# import grpc
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
promAddress    = int(os.getenv('PROM_ADDRESS', '8084'))
videoPubsub    = os.getenv('VIDEO_PUBSUB', 'video-pubsub')
scaleTopic     = os.getenv('SCALE_TOPIC', 'scale')
thumbnailTopic = os.getenv('THUMBNAIL_TOPIC', 'thumbnail')
videoStore     = os.getenv('VIDEO_STORE', 'video-store')
thumbnailStore = os.getenv('THUMBNAIL_STORE', 'thumbnail-store')
numWorkers = int(os.getenv('WORKERS', '10'))
logging.basicConfig(level=logging.INFO)

# prometheus metrics
uploadReq = prometheus_client.Counter(
    'video_frontend_upload_total', 
    'Number of upload requests processed by video-frontend')
infoReq = prometheus_client.Counter(
    'video_frontend_info_total', 
    'Number of info requests processed by video-frontend')
videoReq = prometheus_client.Counter(
    'video_frontend_video_total', 
    'Number of video requests processed by video-frontend')
rateReq = prometheus_client.Counter(
    'video_frontend_rate_total', 
    'Number of rate requests processed by video-frontend')
getRateReq = prometheus_client.Counter(
    'video_frontend_get_rate_total', 
    'Number of get-rate requests processed by video-frontend')
# requests sent by frontend
scaleReq = prometheus_client.Counter(
    'video_frontend_scale_total', 
    'Number of video scale requests processed by video-frontend')
thumbnailReq = prometheus_client.Counter(
    'video_frontend_thumbnail_total', 
    'Number of video thumbnail requests sent by video-frontend')
# todo: the latency range needs refined
uploadLat = prometheus_client.Histogram(
    'video_frontend_upload_lat_hist',
    'Latency (ms) histogram of video-frontend upload requests',
    buckets=pyutil.latBucketsFFprobe()
)
infoLat = prometheus_client.Histogram(
    'video_frontend_info_lat_hist',
    'Latency (ms) histogram of video-frontend info requests',
    buckets=pyutil.latBuckets()
)
videoLat = prometheus_client.Histogram(
    'video_frontend_video_lat_hist',
    'Latency (ms) histogram of video-frontend video requests',
    buckets=pyutil.latBuckets()
)
rateLat = prometheus_client.Histogram(
    'video_frontend_rate_lat_hist',
    'Latency (ms) histogram of video-frontend rate requests',
    buckets=pyutil.latBuckets()
)
getRateLat = prometheus_client.Histogram(
    'video_frontend_get_rate_lat_hist',
    'Latency (ms) histogram of video-frontend get-rate requests',
    buckets=pyutil.latBuckets()
)
uploadStoreLat = prometheus_client.Histogram(
    'video_frontend_upload_store_lat_hist',
    'Latency (ms) histogram of writing video-store (kvs/db) for video-frontend upload requests',
    buckets=pyutil.latBuckets()
)
infoStoreLat = prometheus_client.Histogram(
    'video_frontend_info_store_lat_hist',
    'Latency (ms) histogram of reading video-store (kvs/db) for video-frontend info requests',
    buckets=pyutil.latBuckets()
)
videoStoreLat = prometheus_client.Histogram(
    'video_frontend_video_store_lat_hist',
    'Latency (ms) histogram of reading video-store (kvs/db) for video-frontend video requests',
    buckets=pyutil.latBuckets()
)
e2eUploadLat = prometheus_client.Histogram(
    'e2e_video_upload_lat_hist',
    'End-to-end latency (ms) histogram of upload-video.',
    buckets=pyutil.latBucketsFFprobe()
)
e2eInfoLat = prometheus_client.Histogram(
    'e2e_video_info_lat_hist',
    'End-to-end latency (ms) histogram of get-info.',
    buckets=pyutil.latBuckets()
)
e2eVideoLat = prometheus_client.Histogram(
    'e2e_get_video_lat_hist',
    'End-to-end latency (ms) histogram of get-video.',
    buckets=pyutil.latBuckets()
)
e2eRateLat = prometheus_client.Histogram(
    'e2e_rate_video_lat_hist',
    'End-to-end latency (ms) histogram of rate.',
    buckets=pyutil.latBuckets()
)
e2eGetRateLat = prometheus_client.Histogram(
    'e2e_get_rate_lat_hist',
    'End-to-end latency (ms) histogram of get-rate.',
    buckets=pyutil.latBuckets()
)

dateRegx = re.compile('[0-9]{4}-[0-9]{2}-[0-9]{2}')
# checkDate returns true if given string conforms to requried date format
def checkDate(dstr: str):
    global dateRegx
    return dateRegx.match(dstr) != None

# folders to hold videos
dataDir = Path('/tmp') / 'video'
os.makedirs(str(dataDir), exist_ok=True)
# worker function
def videoProcessor(req):
    t = int(time.time() * 1000)
    # logging.info('At %d videoProcessor receives work: %s' %(t, str(req['tempf'])))
    tempf = req['tempf']
    resp = {
        'succ': False,
        'err': None,
        'probe': None,
    }
    try:
        resp['probe'] = ffmpeg.probe(str(tempf))
        resp['succ'] = True
    except ffmpeg.Error as e:
        resp['err'] = e
        # raise RuntimeError('ffprobe stdout: %s, stderr: %s' %(e.stdout, e.stderr))
    # logging.info('videoProcessor completes work %d: %s, succ=%s' %(
    #     t, str(req['tempf']), str(resp['succ'])))   
    return resp

workerPool = None

MAX_PAYLOAD = 64 * 1024 * 1024 # 64MB
# MAX_PAYLOAD=-1
# server
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

# upload a new video
@app.method(name='upload')
def uploadVideo(request: InvokeMethodRequest) -> InvokeMethodResponse:
    # promethues metrics
    global uploadReq
    global uploadLat
    global uploadStoreLat
    global e2eUploadLat
    global workerPool
    # update req counter
    uploadReq.inc()
    ts = int(time.time() * 1000)
    # dt = datetime.now(timezone.utc)
    dt = datetime.now()
    dt_str = ''
    # parse request
    data = json.loads(request.text())
    user_id = data['user']
    video_b64 = data['video_b64']
    desc = data['description']
    if 'date' in data and data['date'] != '':
        if checkDate(data['date']):
            dt_str = data['date']
        else:
            # return error 
            raise ValueError('Error: invalid date in user request: %s' %data['date'])
    else:
        dt_str = pyutil.dtToDate(dt)
    send_unix_ms = data['send_unix_ms']
    client_unix_ms = send_unix_ms
    if len(video_b64) == 0:
        raise ValueError('Empty video uploaded')
    # decode video data into bytes, run ffprobe to get duration & resolution
    tmp_video_id = pyutil.videoId(user_id, ts, None)
    tempf = dataDir / ('%d-%s' %(getCtr(), tmp_video_id))
    video_bytes = base64.b64decode(video_b64)
    with open(str(tempf), 'wb+') as f:
        f.write(video_bytes)
    # dispatch to worker pool
    work = {
        'tempf': tempf,
    }
    fresult = workerPool.apply_async(videoProcessor, (work,))
    result = fresult.get()
    if not result['succ']:
        e = result['err']
        raise RuntimeError('ffprobe stdout: %s, stderr: %s' %(e.stdout, e.stderr)) 
    # remove temp files
    if os.path.exists(str(tempf)):
        os.remove(str(tempf))
    probe = result['probe']
    dur = float(probe['format']['duration'])
    format = pyutil.pickFormat(probe['format']['format_name'])
    video_id = pyutil.videoId(user_id, ts, format)
    width = None
    height = None
    video_stream = next((stream for stream in probe['streams'] if stream['codec_type'] == 'video'), None)
    if video_stream != None:
        width = video_stream['width']
        height = video_stream['height']
    res = pyutil.widthToResolution(width)
    native_data_id = pyutil.videoDataId(video_id, res)
    avail_reso = pyutil.availResolution(res)
    # target widths
    scale_widths = pyutil.scaleWidth(width)
    # latency metrics
    epoch = time.time()*1000
    serv_lat = epoch - send_unix_ms
    with DaprClient(max_grpc_message_length=MAX_PAYLOAD) as d:
        # save the native video
        d.save_state(
            store_name=videoStore, 
            key=native_data_id, 
            value=video_bytes,
            options=StateOptions(
                consistency=Consistency.strong,
                concurrency=Concurrency.last_write,
            ),
        )
        # video-store latency metric
        uploadStoreLat.observe(time.time()*1000 - epoch)
        epoch = time.time()*1000
        # issue scale requests
        for w in scale_widths:
            h = int(w / width * height)
            if h % 2 == 1:
                h += 1
            scale_req = {
                'video_id': video_id,
                'data_id': native_data_id,
                'width': w,
                'height': h,
                'send_unix_ms': int(time.time()*1000),
                'client_unix_ms': int(client_unix_ms),
            }
            resp = d.publish_event(
                pubsub_name=videoPubsub,
                topic_name=scaleTopic,
                data=json.dumps(scale_req),
                data_content_type='application/json',
            )
            scaleReq.inc()
        # issue thumbnail requests
        if video_stream != None:
            thumbnail_req = {
                'video_id': video_id,
                'data_id': native_data_id,
                'duration': dur,
                'send_unix_ms': int(time.time()*1000),
                'client_unix_ms': int(client_unix_ms),
            }
            resp = d.publish_event(
                pubsub_name=videoPubsub,
                topic_name=thumbnailTopic,
                data=json.dumps(thumbnail_req),
                data_content_type='application/json',
            )
            thumbnailReq.inc()
        # save video meta
        serv_lat += time.time()*1000 - epoch
        meta_req = {
            'video_id': video_id,
            'user_id': user_id,
            'resolutions': avail_reso,
            'duration': dur,
            'date': dt_str,
            'description': desc,
            'send_unix_ms': int(time.time()*1000),
        }
        resp = d.invoke_method(
            'dapr-video-info',
            'upload',
            data=json.dumps(meta_req),
        )
        resp_data = json.loads(resp.text())
        epoch = time.time()*1000
        serv_lat += epoch - resp_data['send_unix_ms']
        # query dates service to add the video to the current date
        dates_req = {
            'date': dt_str,
            'video_id': video_id,
            'send_unix_ms': int(time.time()*1000),
        }
        resp = d.invoke_method(
            'dapr-dates',
            'upload',
            data=json.dumps(dates_req),
        )
        resp_data = json.loads(resp.text())
        # update latency metrics
        epoch = time.time()*1000
        serv_lat += epoch - resp_data['send_unix_ms']
        uploadLat.observe(serv_lat)
        e2eUploadLat.observe(epoch - client_unix_ms)
        resp = {
            'video_id': video_id,
        }
        return InvokeMethodResponse(json.dumps(resp), 'application/json')

# fetching info of specified videos (thumbnails included)
@app.method(name='info')
def getVideoInfo(request: InvokeMethodRequest) -> InvokeMethodResponse:
    # promethues metrics
    global infoReq
    global infoLat
    global infoStoreLat
    global e2eInfoLat
    # update req counter
    infoReq.inc()
    # parse request
    data = json.loads(request.text())
    video_ids = data['videos']
    send_unix_ms = data['send_unix_ms']
    client_unix_ms = send_unix_ms
    # thumbnail ids
    thumbnail_ids = []
    video_info = {}
    for vid in video_ids:
        thumbnail_ids.append(pyutil.thumbnailId(vid))
        video_info[vid] = {}
    # latency metrics
    epoch = time.time()*1000
    serv_lat = epoch - send_unix_ms
    with DaprClient(max_grpc_message_length=MAX_PAYLOAD) as d:
        # fetch the thumbnail images
        items = d.get_bulk_state(
            store_name=thumbnailStore, 
            keys=thumbnail_ids).items
        # video-store latency metric
        infoStoreLat.observe(time.time()*1000 - epoch)
        epoch = time.time()*1000
        # encode image into b64
        for it in items:
            k = it.key
            vid = pyutil.thumbnailToVideo(k)
            if vid not in video_info:
                raise ValueError('Extracted video id %s does not match given videos' %(vid))
            else:
                if isinstance(d, str):
                    video_info[vid]['thumbnail'] = base64.b64encode(it.data.encode('ascii')).decode('ascii')
                else:
                    video_info[vid]['thumbnail'] = base64.b64encode(it.data).decode('ascii')
                # logging.info('thumbnail type = %s' %type(video_info[vid]['thumbnail']))
        # query dapr-video-info to get video info 
        info_req = {
            'video_ids': video_ids,
            'upstream': 'frontend',
            'send_unix_ms': int(time.time()*1000),
        }
        # update latency metric
        serv_lat += time.time()*1000 - epoch
        # call dapr-video-info
        resp = d.invoke_method(
            'dapr-video-info',
            'info',
            data=json.dumps(info_req),
        )
        resp_data = json.loads(resp.text())
        for vid in video_info:
            if vid not in resp_data['info']:
                raise RuntimeError('missing info of video %s' %vid)
            else:
                video_info[vid]['info'] = {
                    'score': resp_data['info'][vid]['rating']['score'],
                    'num': resp_data['info'][vid]['rating']['num'],
                    'views': resp_data['info'][vid]['views'],
                    'meta':  resp_data['info'][vid]['meta'],
                }
        # update latency metric
        epoch = time.time()*1000
        serv_lat += epoch - resp_data['send_unix_ms']
        infoLat.observe(serv_lat)
        e2eInfoLat.observe(epoch - client_unix_ms)
        resp = {
            'video_info': video_info,
        }
        return InvokeMethodResponse(json.dumps(resp), 'application/json')

# fetching actual video data with specified resolution
@app.method(name='video')
def getVideoData(request: InvokeMethodRequest) -> InvokeMethodResponse:
    # promethues metrics
    global videoReq
    global videoLat
    global videoStoreLat
    global e2eVideoLat
    # update req counter
    videoReq.inc()
    # parse request
    data = json.loads(request.text())
    video_id = data['video']
    res = data['resolution']
    data_id = pyutil.videoDataId(video_id=video_id, res=res)
    send_unix_ms = float(data['send_unix_ms'])
    client_unix_ms = send_unix_ms
    # latency metrics
    epoch = time.time()*1000
    serv_lat = epoch - send_unix_ms
    with DaprClient(max_grpc_message_length=MAX_PAYLOAD) as d:
        # call video-info to increment view count
        view_req = {
            'video_id': video_id,
            'send_unix_ms': int(epoch),
        }
        resp = d.invoke_method(
            'dapr-video-info',
            'view',
            data=json.dumps(view_req),
        )
        resp_data = json.loads(resp.text())
        # update latency metric
        epoch = time.time()*1000
        serv_lat += epoch - resp_data['send_unix_ms']
        # fetch the actual video
        state = d.get_state(
            store_name=videoStore, 
            key=data_id)
        # video-store latency metric
        videoStoreLat.observe(time.time()*1000 - epoch)
        epoch = time.time()*1000
        data = state.data
        resp = {}
        # encode video into b64
        if isinstance(data, str):
            resp['data'] = base64.b64encode(data.encode('ascii')).decode('ascii')
        else:
            resp['data'] = base64.b64encode(data).decode('ascii')
        final_epoch = time.time()*1000
        serv_lat += final_epoch - epoch
        videoLat.observe(serv_lat)
        e2eVideoLat.observe(final_epoch - client_unix_ms)
        return InvokeMethodResponse(json.dumps(resp), 'application/json')

# rate (or change rating) of a certain video
@app.method(name='rate')
def rateVideo(request: InvokeMethodRequest) -> InvokeMethodResponse:
    # promethues metrics
    global rateReq
    global rateLat
    global e2eRateLat
    # update req counter
    rateReq.inc()
    # parse request
    data = json.loads(request.text())
    video_id = data['video']
    user_id = data['user']
    comment = data['comment']
    score = data['score']
    send_unix_ms = data['send_unix_ms']
    client_unix_ms = send_unix_ms
    # latency metrics
    epoch = time.time()*1000
    serv_lat = epoch - send_unix_ms
    with DaprClient() as d:
        # call dapr-user-rating to update score & comment of user and get original rating
        ur_req = {
            'user_id': user_id,
            'video_id': video_id,
            'comment': comment,
            'score': score,
            'send_unix_ms': int(epoch),
        }
        resp = d.invoke_method(
            'dapr-user-rating',
            'rate',
            data=json.dumps(ur_req),
        )
        resp_data = json.loads(resp.text())
        rate_exist = resp_data['exist']
        ori_score = resp_data['ori_score']
        # update latency metric
        epoch = time.time()*1000
        serv_lat += epoch - resp_data['send_unix_ms']
        # call dapr-info to update the overall score of the video
        vr_req = {
            'video_id': video_id,
            'change': rate_exist,
            'score': score,
            'ori_score': ori_score,
            'send_unix_ms': int(time.time()*1000),
        }
        resp = d.invoke_method(
            'dapr-video-info',
            'rate',
            data=json.dumps(vr_req),
        )
        resp_data = json.loads(resp.text())
        # update latency metric
        epoch = time.time()*1000
        serv_lat += epoch - resp_data['send_unix_ms']
        rateLat.observe(serv_lat)
        e2eRateLat.observe(epoch - client_unix_ms)
        resp = {
            'video_id': video_id,
        }
        return InvokeMethodResponse(json.dumps(resp), 'application/json')

# get user's rating of a certain movie
@app.method(name='get_rate')
def getVideoRating(request: InvokeMethodRequest) -> InvokeMethodResponse:
    # promethues metrics
    global getRateReq
    global getRateLat
    global e2eGetRateLat
    # update req counter
    getRateReq.inc()
    # parse request
    data = json.loads(request.text())
    video_id = data['video']
    user_id = data['user']
    send_unix_ms = float(data['send_unix_ms'])
    client_unix_ms = send_unix_ms
    # latency metrics
    epoch = time.time()*1000
    serv_lat = epoch - send_unix_ms
    with DaprClient() as d:
        # call dapr-user-rating to update score & comment of user and get original rating
        ur_req = {
            'user_id': user_id,
            'video_id': video_id,
            'send_unix_ms': int(epoch),
        }
        resp = d.invoke_method(
            'dapr-user-rating',
            'get',
            data=json.dumps(ur_req),
        )
        resp_data = json.loads(resp.text())
        rate_exist = resp_data['exist']
        comment = resp_data['comment']
        score = float(resp_data['score'])
        # update latency metric
        epoch = time.time()*1000
        serv_lat += epoch - resp_data['send_unix_ms']
        getRateLat.observe(serv_lat)
        e2eGetRateLat.observe(epoch - client_unix_ms)
        resp = {
            'exist': rate_exist,
            'comment': comment,
            'score': score,
        }
        return InvokeMethodResponse(json.dumps(resp), 'application/json')

if __name__ == '__main__':
    # worker pool
    workerPool = get_context("spawn").Pool(processes=numWorkers)
    # start prometheus
    prometheus_client.start_http_server(promAddress)
    # start the service
    app.run(serviceAddress)