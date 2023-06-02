import os
import json
import logging
import time
# dapr
from dapr.clients import DaprClient
from dapr.ext.grpc import App, InvokeMethodRequest, InvokeMethodResponse

# global variables
serviceAddress = int(os.getenv('ADDRESS', '5005'))
logging.basicConfig(level=logging.INFO)

# server 
app = App()
# handlers
@app.method(name='dummy')
def dummy(request: InvokeMethodRequest) -> InvokeMethodResponse:
    resp = {'texts': 'Hi!'}
    return InvokeMethodResponse(
        data=json.dumps(resp),
        content_type='application/json')

# start the service
app.run(serviceAddress)