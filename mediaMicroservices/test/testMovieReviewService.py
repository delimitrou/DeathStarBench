import sys
sys.path.append('../gen-py')

from media_service import MovieReviewService
from media_service.ttypes import ServiceException

from thrift import Thrift
from thrift.transport import TSocket
from thrift.transport import TTransport
from thrift.protocol import TBinaryProtocol

import random
import string
from time import time

def write_movie_review():
  socket = TSocket.TSocket("ath-8.ece.cornell.edu", 9090)
  transport = TTransport.TFramedTransport(socket)
  protocol = TBinaryProtocol.TBinaryProtocol(transport)
  client = MovieReviewService.Client(protocol)

  transport.open()
  for i in range(101, 200):
    req_id = random.getrandbits(63)
    timestamp = int(time() * 1000)
    movie_num = random.randint(0, 5)
    movie_id = "movie_id_" + str(movie_num)

    client.UploadMovieReview(req_id, movie_id, i, timestamp, {})
  transport.close()

def read_movie_reviews():
  socket = TSocket.TSocket("ath-8.ece.cornell.edu", 9090)
  transport = TTransport.TFramedTransport(socket)
  protocol = TBinaryProtocol.TBinaryProtocol(transport)
  client = MovieReviewService.Client(protocol)

  transport.open()
  for i in range(100):
    req_id = random.getrandbits(63)
    movie_num = random.randint(0, 5)
    movie_id = "movie_id_" + str(movie_num)
    start = random.randint(0, 10)
    stop = start + random.randint(1, 10)

    print(client.ReadMovieReviews(req_id, movie_id, start, stop, {}))
  transport.close()


if __name__ == '__main__':
  try:
    write_movie_review()
    read_movie_reviews()
  except ServiceException as se:
    print('%s' % se.message)
  except Thrift.TException as tx:
    print('%s' % tx.message)