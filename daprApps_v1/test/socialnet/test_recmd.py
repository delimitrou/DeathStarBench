import os
import sys
import requests
import time
import json
import base64
from pathlib import Path
import random

recmd_url = 'http://localhost:31992/v1.0/invoke/dapr-recommend/method/recmd'

from pathlib import Path
img_path = Path(__file__).parent.resolve() / '..' / 'data'
sys.path.append(str(img_path))

# Apart from following each user himself
# tester-0 follows tester-1, tester-2 & tester-3
# tester-1 follows tester-2, tester-3 & tester-4 
# tester-2 follows tester-3, tester-4 & tester-0 
# ...
# tester-4 follows tester-0, tester-1 & tester-2
# 
# tester-0 followed by tester-2, tester-3 & tester-4
# tester-1 followed by tester-3, tester-4 & tester-0 
# ...
# tester-4 followed by tester-1, tester-2 & tester-3
def make_user(i: int):
    return 'tester-' + str(i)

for i in range(0, 5):
    print('------ follow recmd for %s ------' %make_user(i))
    payload = {
        'send_unix_ms': int(time.time() * 1000),
        'user_id': make_user(i),
    }
    r = requests.post(recmd_url, json=payload)
    resp = json.loads(r.text)
    # print(resp)
    assert make_user(i) not in resp['user_ids']
    print(len(resp['user_ids']))
    print(sorted(resp['user_ids']))

