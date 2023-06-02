
#!/bin/bash
cp -r ../../../test/video ./
docker build -t sailresearch/dapr-test-video-scale:latest .
docker push sailresearch/dapr-test-video-scale:latest
rm -rf video