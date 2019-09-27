import sys
sys.path.append('../gen-py')

import random
from media_service import MovieIdService
from media_service.ttypes import ServiceException

from thrift import Thrift
from thrift.transport import TSocket
from thrift.transport import TTransport
from thrift.protocol import TBinaryProtocol

def register_movie():
  socket = TSocket.TSocket("ath-8.ece.cornell.edu", 9090)
  transport = TTransport.TFramedTransport(socket)
  protocol = TBinaryProtocol.TBinaryProtocol(transport)
  client = MovieIdService.Client(protocol)

  transport.open()
  for i in range(100):
    req_id = random.getrandbits(63)
    movie_index = i
    title = "movie_" + str(movie_index)
    movie_id = "movie_id_" + str(movie_index)
    client.RegisterMovieId(req_id, title, movie_id, {})
  transport.close()

def upload_movie_id():
  socket = TSocket.TSocket("ath-8.ece.cornell.edu", 9090)
  transport = TTransport.TFramedTransport(socket)
  protocol = TBinaryProtocol.TBinaryProtocol(transport)
  client = MovieIdService.Client(protocol)

  transport.open()
  for i in range(100):
    req_id = random.getrandbits(63)
    movie_index = random.randint(0, 4)
    title = "movie_" + str(movie_index)
    rating = random.randint(0, 10)
    client.UploadMovieId(req_id, title, rating, {})
  transport.close()

if __name__ == '__main__':
  try:
    # register_movie()
    upload_movie_id()
  except ServiceException as se:
    print('%s' % se.message)
  except Thrift.TException as tx:
    print('%s' % tx.message)