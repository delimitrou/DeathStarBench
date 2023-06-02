from PIL import Image
import io
from pathlib import Path

img_dir = Path(__file__).parent.resolve() / 'data'
images = [
    'panda2.jpg',  
    'panda.jpeg',  
    'shiba2.jpg',  
    'shiba.jpg',
    ]

for img in images:
    with open(str(img_dir / img), 'rb') as f:
        img_data = f.read()
        pil_img = Image.open(io.BytesIO(img_data))
        print(img)
