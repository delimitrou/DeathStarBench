# ml
from transformers import pipeline
sentiment = pipeline('sentiment-analysis')
sentiment('Hello World') # warm up the model