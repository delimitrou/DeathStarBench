import numpy as np

def pickFormat(format: str) -> str:
    if ',' not in format:
        return format
    else:
        all_formats = format.replace(' ', '').split(',')
        if 'mp4' in all_formats and 'mov' in all_formats:
            return 'mp4'
        else:
            return all_formats[0]

# LatBuckets generate a common latency histogram buckets for prometheus histogram
def latBuckets():
    # 1-200
    buckets = list(np.arange(1.0, 201.0, 1.0))
    # 205-500
    buckets += list(np.arange(205.0, 505.0, 5.0))
    # 510-5000
    buckets += list(np.arange(510.0, 5010.0, 10.0))
    # 510-5000
    buckets += list(np.arange(5025.0, 10025.0, 25.0))
	# 10000 - 60000
    buckets += list(np.arange(11000.0, 61000.0, 1000.0))
    return buckets

# LatBucketsMl generate a prometheus latency histogram buckets for ffprobe workloads
def latBucketsLong():
    # 5 - 5005
    buckets = list(np.arange(5.0, 5005.0, 5.0))
    # 5010 - 20010
    buckets += list(np.arange(5010.0, 10010.0, 10.0))
    # 10050 - 60000
    buckets += list(np.arange(10020.0, 60020.0, 20.0))
    # 60200 - 150000
    buckets += list(np.arange(60200.0, 150200.0, 200.0))
    return buckets