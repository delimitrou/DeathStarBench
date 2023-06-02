# Absolute path to this script, e.g. /home/user/bin/foo.sh
SCRIPT=$(readlink -f "$0")
# Absolute path this script is in, thus /home/user/bin
DIR=$(dirname "$SCRIPT")
# echo $SCRIPTPATH
BUILDPATH="$(dirname "$DIR")"
FILE=${DIR}"/Dockerfile"
echo $BUILDPATH
cd $BUILDPATH && docker build -f $FILE -t sailresearch/dapr-video-scale:latest .
docker push sailresearch/dapr-video-scale:latest