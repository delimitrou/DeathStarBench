import os
# import sys
import numpy as np
import json
import logging
import time
# dapr
from dapr.clients import DaprClient
from dapr.ext.grpc import App, InvokeMethodRequest, InvokeMethodResponse
# prometheus
import prometheus_client
# # util
# from pathlib import Path
# util_path = Path(__file__).parent.resolve() / '..' / 'pyutil'
# sys.path.append(str(util_path))
# # warnings
# import warnings
# warnings.filterwarnings("ignore")
# global variables
serviceAddress = int(os.getenv('ADDRESS', '5005'))
promAddress = int(os.getenv('PROM_ADDRESS', '8084'))
logging.basicConfig(level=logging.INFO)

# LatBuckets generate a latency histogram buckets for prometheus histogram
def latBuckets():
    # 1-200
    buckets = list(np.arange(1.0, 201.0, 1.0))
    # 205-500
    buckets += list(np.arange(205.0, 505.0, 5.0))
    # 510-1000
    buckets += list(np.arange(510.0, 1010.0, 10.0))
	# 1100 - 5000
    buckets += list(np.arange(1100.0, 5100.0, 100.0))
	# 10000 - 60000
    buckets += list(np.arange(10000.0, 65000.0, 5000.0))
    return buckets

# prometheus metrics
# todo: the latency range needs refined
promLat = prometheus_client.Histogram(
    'proxy_lat_hist',
    'Latency of proxy request (ms)',
    buckets=latBuckets(),
)

MAX_PAYLOAD = 64 * 1024 * 1024 # 64MB
# MAX_PAYLOAD=-1
grpcOptions = [
    ('grpc.max_send_message_length', MAX_PAYLOAD),
    ('grpc.max_receive_message_length', MAX_PAYLOAD),
]

# server 
app = App()
# handlers
@app.method(name='forward')
def forward(request: InvokeMethodRequest) -> InvokeMethodResponse:
    global promLat
    if request.content_type != 'application/json':
        logging.error('Invalid content type: %s' %request.content_type)
        return InvokeMethodResponse(
            data='Invalid content type: %s' %request.content_type,
            content_type='text/plain')
    req = json.loads(request.text())
    send_unix_ms = float(req['send_unix_ms'])
    # rpc downstream info
    method = req['method']
    del req['method']
    downstream = req['downstream']
    del req['downstream']
    # update timestamp
    epoch = time.time() * 1000
    req['send_unix_ms'] = int(epoch)
    serv_lat = epoch - send_unix_ms
    # logging.info('invoke service: %s, method: %s, payload: %s' %(
    #     downstream, method, json.dumps(req)
    # ))
    with DaprClient(max_grpc_message_length=MAX_PAYLOAD) as d:
        resp = d.invoke_method(
            downstream,
            method,
            data=json.dumps(req),
        )
    is_resp_json = True
    try:
        resp_data = json.loads(resp.text())
    except:
        is_resp_json = False
    if is_resp_json:
        # update latency metrics
        epoch = time.time()*1000
        if 'send_unix_ms' in resp_data:
            serv_lat += epoch - resp_data['send_unix_ms']
            resp_data['send_unix_ms'] = int(epoch)
        promLat.observe(serv_lat)
        return InvokeMethodResponse(
            data=json.dumps(resp_data),
            content_type='application/json')
    else:
        promLat.observe(serv_lat)
        return InvokeMethodResponse(
            data=resp.text(),
            content_type='application/octet-stream')

# start prometheus
prometheus_client.start_http_server(promAddress)

# start the service
app.run(serviceAddress)