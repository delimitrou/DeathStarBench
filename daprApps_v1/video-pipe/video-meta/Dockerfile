FROM ubuntu:20.04

RUN apt update
RUN DEBIAN_FRONTEND=noninteractive TZ=Etc/UTC apt-get -y install tzdata
RUN apt install -y bash \
                   build-essential \
                   git \
                   curl \
                   ca-certificates \
                   ffmpeg \
                   python3-pip && \
    rm -rf /var/lib/apt/lists

# RUN apt update && \
#     apt install -y bash \
#                    build-essential \
#                    git \
#                    curl \
#                    ca-certificates \
#                    ffmpeg \
#                    python3-pip && \
#     rm -rf /var/lib/apt/lists
RUN python3 -m pip install --no-cache-dir --upgrade pip

WORKDIR /app

ADD video-meta/requirements.txt .
RUN python3 -m pip install -r requirements.txt

COPY video-meta /app/video-meta
COPY pyutil /app/pyutil

ENV PYTHONPATH "/app"
RUN mkdir -p /tmp/prom_multiproc
ENV PROMETHEUS_MULTIPROC_DIR "/tmp/prom_multiproc"

CMD [ "python3", "video-meta/server.py" ]