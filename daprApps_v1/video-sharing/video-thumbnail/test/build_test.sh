
#!/bin/bash
cp -r ../../../test/video ./video
docker build -t sailresearch/dapr-test-video-thumbnail:latest .
docker push sailresearch/dapr-test-video-thumbnail:latest
rm -rf video