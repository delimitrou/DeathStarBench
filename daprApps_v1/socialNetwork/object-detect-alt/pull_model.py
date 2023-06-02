# ml
from transformers import pipeline
from PIL import Image
import numpy as np

data = np.random.randint(0,255,(100,100,3), dtype='uint8')
img = Image.fromarray(data,'RGB')

# objectDetector = pipeline('object-detection')
print('hf-internal-testing/tiny-detr-mobilenetsv3')
objectDetector = pipeline(model='hf-internal-testing/tiny-detr-mobilenetsv3')
objectDetector([img]) # warm up the model