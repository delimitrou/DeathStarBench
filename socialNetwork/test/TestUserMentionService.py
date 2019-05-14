import sys
sys.path.append('../gen-py')

import uuid
from social_network import UserMentionService

from thrift import Thrift
from thrift.transport import TSocket
from thrift.transport import TTransport
from thrift.protocol import TBinaryProtocol

def main():
  socket = TSocket.TSocket("ath-8.ece.cornell.edu", 9090)
  transport = TTransport.TFramedTransport(socket)
  protocol = TBinaryProtocol.TBinaryProtocol(transport)
  client = UserMentionService.Client(protocol)

  transport.open()
  req_id = uuid.uuid4().int & 0X7FFFFFFFFFFFFFFF

  user_mentions = ["username_0", "username_1", "username_2"]

  print(client.UploadUserMentions(req_id, user_mentions, {}))
  transport.close()

if __name__ == '__main__':
  try:
    main()
  except Thrift.TException as tx:
    print('%s' % tx.message)