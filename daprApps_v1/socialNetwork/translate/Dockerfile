FROM sailresearch/transformers-dapr:latest

WORKDIR /app

ADD translate/requirements.txt .
RUN python3.7 -m pip install -r requirements.txt

COPY translate /app/translate
COPY pyutil /app/pyutil
RUN python3.7 translate/pull_model.py

ENV PYTHONPATH "/app"

CMD [ "python3.7", "translate/server.py" ]