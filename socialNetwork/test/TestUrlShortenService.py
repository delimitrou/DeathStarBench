import sys
sys.path.append('../gen-py')

import uuid
from social_network import UrlShortenService
from social_network.ttypes import Url

from thrift import Thrift
from thrift.transport import TSocket
from thrift.transport import TTransport
from thrift.protocol import TBinaryProtocol

def main():
  socket = TSocket.TSocket("ath-8.ece.cornell.edu", 9090)
  transport = TTransport.TFramedTransport(socket)
  protocol = TBinaryProtocol.TBinaryProtocol(transport)
  client = UrlShortenService.Client(protocol)

  transport.open()
  req_id = uuid.uuid4().int & ( 1 << 32 )

  urls = ["https://url_0.com", "https://url_1.com", "https://url_2.com"]

  print(client.UploadUrls(req_id, urls, {}))
  transport.close()

if __name__ == '__main__':
  try:
    main()
  except Thrift.TException as tx:
    print('%s' % tx.message)