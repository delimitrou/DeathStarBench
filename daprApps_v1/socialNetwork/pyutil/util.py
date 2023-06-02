import numpy as np

# pubsub reliver interval in ms
def redeliverInterval():
    return 60000

# LatBuckets generate a common latency histogram buckets for prometheus histogram
def latBuckets():
    # 1-100
    buckets = list(np.arange(1.0, 101.0, 1.0))
    # 105-500
    buckets += list(np.arange(105.0, 505.0, 5.0))
    # 510-1000
    buckets += list(np.arange(510.0, 1010.0, 10.0))
	# 1050 - 2500
    buckets += list(np.arange(1050.0, 2550.0, 50.0))
	# 2600 - 5000
    buckets += list(np.arange(2600.0, 5100.0, 100.0))
	# 10000 - 60000
    buckets += list(np.arange(10000.0, 65000.0, 5000.0))
    return buckets

# LatBucketsMl generate a prometheus latency histogram buckets for ml workloads
def latBucketsMl():
    # 10 - 1000
    buckets = list(np.arange(10.0, 1010.0, 10.0))
    # 1025 - 2500
    buckets += list(np.arange(1025.0, 2525.0, 25.0))
    # 2550 - 5000
    buckets += list(np.arange(2550.0, 5050.0, 50.0))
    # 10000 - 60000
    buckets += list(np.arange(10000.0, 65000.0, 5000.0))
    return buckets

# LatBucketsMl generate a prometheus latency histogram buckets for long running ml workloads
def latBucketsLongMl():
    # 25 - 5000
    buckets = list(np.arange(25.0, 5025.0, 25.0))
    # 5000 - 10000
    buckets += list(np.arange(5050.0, 10050.0, 50.0))
    # 10000 - 30000
    buckets += list(np.arange(10100.0, 30100.0, 100.0))
    # 31000 - 60000
    buckets += list(np.arange(31000.0, 61000.0, 1000.0))
    return buckets