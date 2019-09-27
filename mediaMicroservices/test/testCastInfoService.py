import sys
sys.path.append('../gen-py')

import random
from media_service import CastInfoService
from media_service.ttypes import ServiceException

from thrift import Thrift
from thrift.transport import TSocket
from thrift.transport import TTransport
from thrift.protocol import TBinaryProtocol


def wrtie_cast_info():
  socket = TSocket.TSocket("ath-8.ece.cornell.edu", 10010)
  transport = TTransport.TFramedTransport(socket)
  protocol = TBinaryProtocol.TBinaryProtocol(transport)
  client = CastInfoService.Client(protocol)

  transport.open()
  for i in range(100):
    req_id = random.getrandbits(63)
    cast_id = i
    name = "name_" + str(i)
    gender = random.randint(0, 1)
    intro = "intro_" + str(i)
    client.WriteCastInfo(req_id, cast_id, name, gender, intro, {})
  transport.close()

def read_cast_info():
  socket = TSocket.TSocket("ath-8.ece.cornell.edu", 10010)
  transport = TTransport.TFramedTransport(socket)
  protocol = TBinaryProtocol.TBinaryProtocol(transport)
  client = CastInfoService.Client(protocol)

  transport.open()
  for i in range(100):
    req_id = random.getrandbits(63)
    cast_ids = set()
    for j in range(10):
      cast_ids.add(random.randint(0, 99))
    print(client.ReadCastInfo(req_id, cast_ids, {}))
  transport.close()

if __name__ == '__main__':
  try:
    wrtie_cast_info()
    read_cast_info()
  except ServiceException as se:
    print('%s' % se.message)
  except Thrift.TException as tx:
    print('%s' % tx.message)