import sys
sys.path.append('../gen-py')

from media_service import UserReviewService
from media_service.ttypes import ServiceException

from thrift import Thrift
from thrift.transport import TSocket
from thrift.transport import TTransport
from thrift.protocol import TBinaryProtocol

import random
from time import time

def write_user_review():
  socket = TSocket.TSocket("ath-8.ece.cornell.edu", 9090)
  transport = TTransport.TFramedTransport(socket)
  protocol = TBinaryProtocol.TBinaryProtocol(transport)
  client = UserReviewService.Client(protocol)

  transport.open()
  for i in range(0, 100):
    req_id = random.getrandbits(63)
    timestamp = int(time() * 1000)
    user_id = random.randint(0, 5)
    client.UploadUserReview(req_id, user_id, i, timestamp, {})
  transport.close()

def read_user_reviews():
  socket = TSocket.TSocket("ath-8.ece.cornell.edu", 9090)
  transport = TTransport.TFramedTransport(socket)
  protocol = TBinaryProtocol.TBinaryProtocol(transport)
  client = UserReviewService.Client(protocol)

  transport.open()
  for i in range(100):
    req_id = random.getrandbits(63)
    user_id = random.randint(0, 5)
    start = random.randint(0, 10)
    stop = start + random.randint(1, 10)

    print(client.ReadUserReviews(req_id, user_id, start, stop, {}))
  transport.close()


if __name__ == '__main__':
  try:
    write_user_review()
    read_user_reviews()
  except ServiceException as se:
    print('%s' % se.message)
  except Thrift.TException as tx:
    print('%s' % tx.message)