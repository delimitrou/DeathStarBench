FROM ubuntu:18.04

RUN apt update && \
    apt install -y bash \
                   build-essential \
                   git \
                   curl \
                   ca-certificates \
                   ffmpeg \
                   python3.7 \
                   python3-pip && \
    rm -rf /var/lib/apt/lists
RUN python3.7 -m pip install --no-cache-dir --upgrade pip

WORKDIR /app

ADD proxy/requirements.txt .
RUN python3.7 -m pip install -r requirements.txt

COPY proxy /app/proxy
# COPY pyutil /app/pyutil

ENV PYTHONPATH "/app"

CMD [ "python3.7", "proxy/server.py" ]