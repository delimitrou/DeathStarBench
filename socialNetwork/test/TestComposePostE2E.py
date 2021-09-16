import sys
sys.path.append('../gen-py')

import uuid
from social_network import TextService
from social_network import MediaService
from social_network import UniqueIdService
from social_network import UserService
from social_network.ttypes import PostType

from thrift import Thrift
from thrift.transport import TSocket
from thrift.transport import TTransport
from thrift.protocol import TBinaryProtocol

def main():
  req_id = uuid.uuid4().int & 0x7FFFFFFFFFFFFFFF
  # text = "text https://url_0 https://url_1 @username_2 https://url_2"
  text = "text"
  post_tyoe = PostType.POST
  media_types = ["png", "png", "png", "png"]
  media_ids = [1, 2, 3, 4]
  creator = "username_0"

  text_socket = TSocket.TSocket("ath-8.ece.cornell.edu", 10007)
  text_transport = TTransport.TFramedTransport(text_socket)
  text_protocol = TBinaryProtocol.TBinaryProtocol(text_transport)
  text_client = TextService.Client(text_protocol)
  text_transport.open()
  text_client.UploadText(req_id, text, {})
  text_transport.close()

  media_socket = TSocket.TSocket("ath-8.ece.cornell.edu", 10006)
  media_transport = TTransport.TFramedTransport(media_socket)
  media_protocol = TBinaryProtocol.TBinaryProtocol(media_transport)
  media_client = MediaService.Client(media_protocol)
  media_transport.open()
  print(media_client.UploadMedia(req_id, media_types, media_ids, {}))
  media_transport.close()

  user_socket = TSocket.TSocket("ath-8.ece.cornell.edu", 10005)
  user_transport = TTransport.TFramedTransport(user_socket)
  user_protocol = TBinaryProtocol.TBinaryProtocol(user_transport)
  user_client = UserService.Client(user_protocol)
  user_transport.open()
  user_client.UploadCreatorWithUsername(req_id, creator, {})
  user_transport.close()

  post_id_socket = TSocket.TSocket("ath-8.ece.cornell.edu", 10008)
  post_id_transport = TTransport.TFramedTransport(post_id_socket)
  post_id_protocol = TBinaryProtocol.TBinaryProtocol(post_id_transport)
  post_id_client = UniqueIdService.Client(post_id_protocol)
  post_id_transport.open()
  post_id_client.UploadUniqueId(req_id, post_tyoe, {})
  post_id_transport.close()



if __name__ == '__main__':
  try:
    main()
  except Thrift.TException as tx:
    print('%s' % tx.message)