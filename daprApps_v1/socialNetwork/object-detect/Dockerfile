FROM sailresearch/transformers-dapr:latest

WORKDIR /app

ADD object-detect/requirements.txt .
RUN python3 -m pip install -r requirements.txt

COPY object-detect /app/object-detect
COPY pyutil /app/pyutil
RUN python3 object-detect/pull_model.py

ENV PYTHONPATH "/app"

CMD [ "python3", "object-detect/server.py" ]