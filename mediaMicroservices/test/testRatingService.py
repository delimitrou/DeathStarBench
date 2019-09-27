import sys
sys.path.append('../gen-py')

from media_service import RatingService
from media_service.ttypes import ServiceException

from thrift import Thrift
from thrift.transport import TSocket
from thrift.transport import TTransport
from thrift.protocol import TBinaryProtocol

import random
import string

def upload_rating():
  socket = TSocket.TSocket("ath-8.ece.cornell.edu", 9090)
  transport = TTransport.TFramedTransport(socket)
  protocol = TBinaryProtocol.TBinaryProtocol(transport)
  client = RatingService.Client(protocol)

  transport.open()
  for i in range (1, 100):
    req_id = random.getrandbits(63)
    movie_id = "movie_id_" + str(random.randint(0, 4))
    rating = random.randint(0, 10)
    client.UploadRating(req_id, movie_id, rating, {})

  transport.close()


if __name__ == '__main__':
  try:
    upload_rating()
  except ServiceException as se:
    print('%s' % se.message)
  except Thrift.TException as tx:
    print('%s' % tx.message)