import sys
sys.path.append('../gen-py')

import random
from social_network import PostStorageService
from social_network.ttypes import Media
from social_network.ttypes import PostType
from social_network.ttypes import Creator
from social_network.ttypes import Url
from social_network.ttypes import UserMention
from social_network.ttypes import Post
from social_network.ttypes import ServiceException

from thrift import Thrift
from thrift.transport import TSocket
from thrift.transport import TTransport
from thrift.protocol import TBinaryProtocol

def store_post():
  socket = TSocket.TSocket("ath-8.ece.cornell.edu", 9090)
  transport = TTransport.TFramedTransport(socket)
  protocol = TBinaryProtocol.TBinaryProtocol(transport)
  client = PostStorageService.Client(protocol)

  transport.open()
  req_id = random.getrandbits(63)
  text = "HelloWorld"
  media_0 = Media(media_id=0, media_type=MediaType.PHOTO)
  media_1 = Media(media_id=1, media_type=MediaType.PHOTO)
  media = [media_0, media_1]
  post_id = 0
  post_type = PostType.POST
  creator = Creator(username="user_0", user_id=0)
  url_0 = Url(shortened_url="shortened_url_0", expanded_url="expanded_url_0")
  url_1 = Url(shortened_url="shortened_url_1", expanded_url="expanded_url_1")
  urls = [url_0, url_1]
  user_mention_0 = UserMention(user_id=1, username="user_1")
  user_mention_1 = UserMention(user_id=2, username="user_2")

  user_mentions = [user_mention_0 ,user_mention_1]
  post = Post(user_mentions=user_mentions, req_id=req_id, creator=creator,
    post_type=post_type, urls=urls, media=media, post_id=post_id,
    text=text)
  client.StorePost(req_id, post, {})
  transport.close()

def read_post():
  socket = TSocket.TSocket("ath-8.ece.cornell.edu", 9090)
  transport = TTransport.TFramedTransport(socket)
  protocol = TBinaryProtocol.TBinaryProtocol(transport)
  client = PostStorageService.Client(protocol)

  transport.open()
  req_id = random.getrandbits(63)
  post_id = 0
  post = client.ReadPost(req_id, post_id, {})
  print(post)

def read_posts():
  socket = TSocket.TSocket("ath-8.ece.cornell.edu", 9090)
  transport = TTransport.TFramedTransport(socket)
  protocol = TBinaryProtocol.TBinaryProtocol(transport)
  client = PostStorageService.Client(protocol)

  transport.open()
  req_id = random.getrandbits(63)
  post_id = [783799431931478016, 783799431931547648, 783799431931551744]
  posts = client.ReadPosts(req_id, post_id, {})
  print(posts)


if __name__ == '__main__':
  try:
    read_posts()
  except ServiceException as se:
    print('%s' % se.message)
  except Thrift.TException as tx:
    print('%s' % tx.message)