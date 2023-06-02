FROM ubuntu:20.04

RUN apt update && \
    apt install -y bash \
                   build-essential \
                   git \
                   curl \
                   ca-certificates \
                   ffmpeg \
                   python3-pip && \
    rm -rf /var/lib/apt/lists
RUN python3 -m pip install --no-cache-dir --upgrade pip

WORKDIR /app

ADD video-scale/requirements.txt .
RUN python3 -m pip install -r requirements.txt

COPY video-scale /app/video-scale
COPY pyutil /app/pyutil

ENV PYTHONPATH "/app"

CMD [ "python3", "video-scale/server.py" ]