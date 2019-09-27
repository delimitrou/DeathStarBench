import sys
sys.path.append('../gen-py')

from media_service import UserService
from media_service import TextService
from media_service import RatingService
from media_service import UniqueIdService
from media_service import MovieIdService
from media_service.ttypes import ServiceException

from thrift import Thrift
from thrift.transport import TSocket
from thrift.transport import TTransport
from thrift.protocol import TBinaryProtocol

import random
import string
from multiprocessing import Process


def register_movies():
  movie_id_socket = TSocket.TSocket("ath-8.ece.cornell.edu", 10002)
  movie_id_transport = TTransport.TFramedTransport(movie_id_socket)
  movie_id_protocol = TBinaryProtocol.TBinaryProtocol(movie_id_transport)
  movie_id_client = MovieIdService.Client(movie_id_protocol)

  movie_id_transport.open()
  for i in range(100):
    req_id = random.getrandbits(63)
    title = "movie_title_" + str(i)
    movie_id = "movie_id_" + str(i)
    movie_id_client.RegisterMovieId(req_id, title, movie_id, {})
  movie_id_transport.close()

def register_users():
  user_socket = TSocket.TSocket("ath-8.ece.cornell.edu", 10005)
  user_transport = TTransport.TFramedTransport(user_socket)
  user_protocol = TBinaryProtocol.TBinaryProtocol(user_transport)
  user_client = UserService.Client(user_protocol)
  user_transport.open()
  for i in range(100):
    req_id = random.getrandbits(63)
    first_name = "first_" + str(i)
    last_name = "last_" + str(i)
    username = "username_" + str(i)
    password = "password_" + str(i)
    user_client.RegisterUserWithId(req_id, first_name, last_name, username, password, i, {})
  user_transport.close()

def worker():
  # text_socket = TSocket.TSocket("text-service", 9090)
  # text_transport = TTransport.TFramedTransport(text_socket)
  # text_protocol = TBinaryProtocol.TBinaryProtocol(text_transport)
  # text_client = TextService.Client(text_protocol)
  #
  # unique_id_socket = TSocket.TSocket("unique-id-service", 9090)
  # unique_id_transport = TTransport.TFramedTransport(unique_id_socket)
  # unique_id_protocol = TBinaryProtocol.TBinaryProtocol(unique_id_transport)
  # unique_id_client = UniqueIdService.Client(unique_id_protocol)
  #
  # rating_socket = TSocket.TSocket("rating-service", 9090)
  # rating_transport = TTransport.TFramedTransport(rating_socket)
  # rating_protocol = TBinaryProtocol.TBinaryProtocol(rating_transport)
  # rating_client = RatingService.Client(rating_protocol)
  #
  # movie_id_socket = TSocket.TSocket("movie-id-service", 9090)
  # movie_id_transport = TTransport.TFramedTransport(movie_id_socket)
  # movie_id_protocol = TBinaryProtocol.TBinaryProtocol(movie_id_transport)
  # movie_id_client = MovieIdService.Client(movie_id_protocol)
  #
  # user_socket = TSocket.TSocket("user-service", 9090)
  # user_transport = TTransport.TFramedTransport(user_socket)
  # user_protocol = TBinaryProtocol.TBinaryProtocol(user_transport)
  # user_client = UserService.Client(user_protocol)

  text_socket = TSocket.TSocket("ath-8.ece.cornell.edu", 10003)
  text_transport = TTransport.TFramedTransport(text_socket)
  text_protocol = TBinaryProtocol.TBinaryProtocol(text_transport)
  text_client = TextService.Client(text_protocol)

  unique_id_socket = TSocket.TSocket("ath-8.ece.cornell.edu", 10001)
  unique_id_transport = TTransport.TFramedTransport(unique_id_socket)
  unique_id_protocol = TBinaryProtocol.TBinaryProtocol(unique_id_transport)
  unique_id_client = UniqueIdService.Client(unique_id_protocol)

  movie_id_socket = TSocket.TSocket("ath-8.ece.cornell.edu", 10002)
  movie_id_transport = TTransport.TFramedTransport(movie_id_socket)
  movie_id_protocol = TBinaryProtocol.TBinaryProtocol(movie_id_transport)
  movie_id_client = MovieIdService.Client(movie_id_protocol)

  user_socket = TSocket.TSocket("ath-8.ece.cornell.edu", 10005)
  user_transport = TTransport.TFramedTransport(user_socket)
  user_protocol = TBinaryProtocol.TBinaryProtocol(user_transport)
  user_client = UserService.Client(user_protocol)

  text_transport.open()
  unique_id_transport.open()
  movie_id_transport.open()
  user_transport.open()

  for i in range(100):
    req_id = random.getrandbits(63)
    user_id = random.randint(0, 99)
    movie_num = random.randint(0, 99)
    rating = random.randint(0, 10)
    text = ''.join(random.choices(string.ascii_lowercase + string.digits, k=256))
    title = "movie_title_" + str(movie_num)

    unique_id_client.UploadUniqueId(req_id, {})
    user_client.UploadUserWithUserId(req_id, user_id, {})
    text_client.UploadText(req_id, text, {})
    movie_id_client.UploadMovieId(req_id, title, rating, {})

  text_transport.close()
  unique_id_transport.close()
  movie_id_transport.close()
  user_transport.close()

def main():
  register_movies()
  register_users()
  processes = []
  for i in range(1):
    p = Process(target=worker)
    processes.append(p)
    p.start()

  for i in range(1):
    processes[i].join()
  print("finished")

if __name__ == '__main__':
  try:
    main()
  except ServiceException as se:
    print('%s' % se.message)
  except Thrift.TException as tx:
    print('%s' % tx.message)