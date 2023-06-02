# ml
from transformers import pipeline
translator = pipeline('translation_en_to_de')
translator('Hello World') # warm up the model