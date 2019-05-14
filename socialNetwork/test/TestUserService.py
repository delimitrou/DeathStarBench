import sys
sys.path.append('../gen-py')

import uuid
from social_network import UserService
from social_network.ttypes import ServiceException

from thrift import Thrift
from thrift.transport import TSocket
from thrift.transport import TTransport
from thrift.protocol import TBinaryProtocol

def register():
  socket = TSocket.TSocket("ath-8.ece.cornell.edu", 10005)
  transport = TTransport.TFramedTransport(socket)
  protocol = TBinaryProtocol.TBinaryProtocol(transport)
  client = UserService.Client(protocol)
  transport.open()
  req_id = uuid.uuid4().int & 0x7FFFFFFFFFFFFFFF
  client.RegisterUser(req_id, "first_name_0", "last_name_0",
    "username_11114", "password_0", {})
  client.RegisterUserWithId(req_id, "first_name_1", "last_name_1",
    "username_1", "password_1", 1, {})
  client.RegisterUserWithId(req_id, "first_name_2", "last_name_2",
    "username_2", "password_2", 2, {})
  transport.close()

def login():
  socket = TSocket.TSocket("ath-8.ece.cornell.edu", 10005)
  transport = TTransport.TFramedTransport(socket)
  protocol = TBinaryProtocol.TBinaryProtocol(transport)
  client = UserService.Client(protocol)
  transport.open()
  req_id = uuid.uuid4().int & 0x7FFFFFFFFFFFFFFF
  print(client.Login(req_id, "username_0", "password_0", {}))
  # print(client.Login(req_id, "username_1", "password_2", {}))
  # print(client.Login(req_id, "username_2", "password_2", {}))
  transport.close()

if __name__ == '__main__':
  try:
    login()
  except ServiceException as se:
    print('%s' % se.message)
  except Thrift.TException as tx:
    print('%s' % tx.message)