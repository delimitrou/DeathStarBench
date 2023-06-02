
#!/bin/bash
cp -r ../../../test/data ./images
docker build -t sailresearch/dapr-test-obj-detect:latest .
docker push sailresearch/dapr-test-obj-detect:latest
rm -rf images