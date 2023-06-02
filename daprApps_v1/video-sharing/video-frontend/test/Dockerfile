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

ADD requirements.txt .
RUN python3.7 -m pip install -r requirements.txt

COPY *.py /app/
COPY video/ /app/video/

CMD [ "python3.7", "dummy_server.py" ]
# CMD [ "tail", "-f", "/dev/null" ]