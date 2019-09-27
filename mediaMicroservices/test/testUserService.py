import sys
sys.path.append('../gen-py')

from media_service import UserService
from media_service import ttypes

from thrift import Thrift
from thrift.transport import TSocket
from thrift.transport import TTransport
from thrift.protocol import TBinaryProtocol

import random
import string
import uuid

def main():
    socket = TSocket.TSocket("ath-8.ece.cornell.edu", 9090)
    transport = TTransport.TFramedTransport(socket)
    protocol = TBinaryProtocol.TBinaryProtocol(transport)
    client = UserService.Client(protocol)

    # transport.open()
    # for i in range (2, 10):
    #     req_id = uuid.uuid4().int & (1<<32)
    #     first_name = "first_name_" + str(i)
    #     last_name = "last_name_" + str(i)
    #     username = "username" + str(i)
    #     password = "password" + str(i)
    #     client.RegisterUser(req_id, first_name, last_name, username, password, {"":""})
    #     print(client.Login(req_id, username, password, {"":""}))
    #
    #
    # transport.close()

    transport.open()
    req_id = uuid.uuid4().int & (1<<32)
    username = "username" + str(1)
    password = "password" + str(1)
    try:
        client.UploadUserWithUsername(req_id, username, {"":""})
    except ttypes.ServiceException as se:
        print('%s' % se.message)
    transport.close()


if __name__ == '__main__':
    try:
        main()
    except Thrift.TException as tx:
        print('%s' % tx.message)