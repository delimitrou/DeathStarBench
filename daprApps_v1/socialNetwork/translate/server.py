import os
import sys
import json
import logging
import time
import numpy as np
# dapr
# from dapr.clients import DaprClient
from dapr.ext.grpc import App, InvokeMethodRequest, InvokeMethodResponse
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
logging.basicConfig(level=logging.INFO)

# ml model
s0 = time.time()
translator = pipeline('translation_en_to_de')
logging.info(translator('Hello World')) # warm up the model
logging.info('Model download time: %.3fs' %(time.time() - s0))

# prometheus metrics
promReq = prometheus_client.Counter(
    'transl_processed', 
    'Number of translation requests processed')
# todo: the latency range needs refined
promLat = prometheus_client.Histogram(
    'transl_lat_lat_hist',
    'Total latency of translation request (ms)',
    buckets=pyutil.latBucketsLongMl(),
)

# server 
app = App()
# handlers
@app.method(name='en_to_de')
def transEnToDe(request: InvokeMethodRequest) -> InvokeMethodResponse:
    global translator
    global promReq
    global promLat
    if 'application/json' not in request.content_type:
        logging.error('Invalid content type: %s' %request.content_type)
        return InvokeMethodResponse(
            data='Invalid content type: %s' %request.content_type,
            content_type='text/plain')
    promReq.inc()
    req = json.loads(request.text())
    de_text = translator(req['text'])
    send_unix_ms = float(req['send_unix_ms'])
    cur_unix_ms = time.time() * 1000
    resp = {
        'translation': de_text,
        'send_unix_ms': int(cur_unix_ms),
    }
    logging.info('en: %s, de: %s, lat: %.3fms' %(
        req['text'], de_text, cur_unix_ms - send_unix_ms))
    promLat.observe(cur_unix_ms - send_unix_ms)
    return InvokeMethodResponse(
        data=json.dumps(resp),
        content_type='application/json')

# start prometheus
prometheus_client.start_http_server(promAddress)

# start the service
app.run(serviceAddress)