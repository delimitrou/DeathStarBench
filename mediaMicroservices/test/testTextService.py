import sys
sys.path.append('../gen-py')

from media_service import TextService

from thrift import Thrift
from thrift.transport import TSocket
from thrift.transport import TTransport
from thrift.protocol import TBinaryProtocol

import random
import string

def main():
    # Make socket
    socket = TSocket.TSocket("ath-8.ece.cornell.edu", 9090)

    # Buffering is critical. Raw sockets are very slow
    transport = TTransport.TFramedTransport(socket)

    # Wrap in a protocol
    protocol = TBinaryProtocol.TBinaryProtocol(transport)

    # Create a client to use the protocol encoder
    client = TextService.Client(protocol)

    # Connect!


    transport.open()
    for i in range (1, 2):
        req_id = random.getrandbits(64) - 2**63
        text = ''.join(random.choices(string.ascii_lowercase + string.digits, k=128))
        client.UploadText(req_id, text)

    transport.close()


if __name__ == '__main__':
    try:
        main()
    except Thrift.TException as tx:
        print('%s' % tx.message)