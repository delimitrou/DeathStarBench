import os
import base64
from pathlib import Path

img_path = Path(__file__).parent.resolve() / 'data'
b64_img_path = Path(__file__).parent.resolve() / 'b64_data'

for fn in os.listdir(str(img_path)):
    with open(str(img_path / fn), 'rb') as f:
        d = f.read()
        b64_d = base64.b64encode(d)
        b64_fn = 'b64_%s' %fn
        with open(str(b64_img_path / b64_fn), 'wb+') as bf:
            bf.write(b64_d)