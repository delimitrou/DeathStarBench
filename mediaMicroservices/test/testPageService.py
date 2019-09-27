import sys
sys.path.append('../gen-py')

import random
from media_service import PageService
from media_service.ttypes import ServiceException

from thrift import Thrift
from thrift.transport import TSocket
from thrift.transport import TTransport
from thrift.protocol import TBinaryProtocol

def read_page():
  socket = TSocket.TSocket("ath-8.ece.cornell.edu", 9090)
  transport = TTransport.TFramedTransport(socket)
  protocol = TBinaryProtocol.TBinaryProtocol(transport)
  client = PageService.Client(protocol)

  transport.open()
  for i in range(100):
    req_id = random.getrandbits(63)
    movie_id = "movie_id_" + str(i)
    print(client.ReadPage(req_id, movie_id, 0, 10, {}))
  transport.close()

if __name__ == '__main__':
  try:
    read_page()
  except ServiceException as se:
    print('%s' % se.message)
  except Thrift.TException as tx:
    print('%s' % tx.message)