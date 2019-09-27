import sys
sys.path.append('../gen-py')

from media_service import ReviewStorageService
from media_service.ttypes import Review
from media_service.ttypes import ServiceException

from thrift import Thrift
from thrift.transport import TSocket
from thrift.transport import TTransport
from thrift.protocol import TBinaryProtocol

import random
import string
from time import time

def write_review():
  socket = TSocket.TSocket("ath-8.ece.cornell.edu", 10007)
  transport = TTransport.TFramedTransport(socket)
  protocol = TBinaryProtocol.TBinaryProtocol(transport)
  client = ReviewStorageService.Client(protocol)

  transport.open()
  for i in range(100):
    req_id = random.getrandbits(63)
    timestamp = int(time())
    review_id = i
    movie_num = random.randint(0, 99)
    user_id = random.randint(0, 99)
    rating = random.randint(0, 10)
    text = ''.join(random.choices(string.ascii_lowercase + string.digits, k=256))
    movie_id = "movie_id_" + str(movie_num)

    review = Review()
    review.req_id = req_id
    review.user_id = user_id
    review.review_id = review_id
    review.text = text
    review.movie_id = movie_id
    review.rating = rating
    review.timestamp = timestamp

    client.StoreReview(req_id, review, {})

  transport.close()

def read_reviews():
  socket = TSocket.TSocket("ath-8.ece.cornell.edu", 10007)
  transport = TTransport.TFramedTransport(socket)
  protocol = TBinaryProtocol.TBinaryProtocol(transport)
  client = ReviewStorageService.Client(protocol)

  transport.open()
  for i in range(100):
    req_id = random.randint(0, 99)
    review_ids = set()
    for j in range(10):
      review_ids.add(random.randint(0, 99))
    print(client.ReadReviews(req_id, review_ids, {}))
  transport.close()

if __name__ == '__main__':
  try:
    write_review()
    read_reviews()
  except ServiceException as se:
    print('%s' % se.message)
  except Thrift.TException as tx:
    print('%s' % tx.message)