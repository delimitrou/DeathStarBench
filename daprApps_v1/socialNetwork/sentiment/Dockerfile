FROM sailresearch/transformers-dapr:latest

WORKDIR /app

ADD sentiment/requirements.txt .
RUN python3 -m pip install -r requirements.txt

COPY sentiment /app/sentiment
COPY pyutil /app/pyutil
RUN python3 sentiment/pull_model.py

ENV PYTHONPATH "/app"

CMD [ "python3", "sentiment/server.py" ]