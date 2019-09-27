import sys
sys.path.append('../gen-py')

from media_service import UniqueIdService

from thrift import Thrift
from thrift.transport import TSocket
from thrift.transport import TTransport
from thrift.protocol import TBinaryProtocol

import random
import string

def main():
    socket = TSocket.TSocket("ath-8.ece.cornell.edu", 9090)
    transport = TTransport.TFramedTransport(socket)
    protocol = TBinaryProtocol.TBinaryProtocol(transport)
    client = UniqueIdService.Client(protocol)

    transport.open()
    for i in range (1, 100) :
        req_id = random.getrandbits(64) - 2**63
        client.UploadUniqueId(req_id)

    transport.close()


if __name__ == '__main__':
    try:
        main()
    except Thrift.TException as tx:
        print('%s' % tx.message)