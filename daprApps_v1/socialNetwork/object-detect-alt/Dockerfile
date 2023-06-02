FROM sailresearch/transformers-dapr:latest

WORKDIR /app

ADD object-detect-alt/requirements.txt .
RUN python3 -m pip install -r requirements.txt

COPY object-detect-alt /app/object-detect
COPY pyutil /app/pyutil
RUN python3 object-detect/pull_model.py

ENV PYTHONPATH "/app"

CMD [ "python3", "object-detect/server.py" ]