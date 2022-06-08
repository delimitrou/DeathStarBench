import aiohttp
import asyncio
import os
import string
import random
import argparse


async def upload_follow(session, addr, user_0, user_1):
  payload = {'user_name': 'username_' + user_0,
             'followee_name': 'username_' + user_1}
  async with session.post(addr + '/wrk2-api/user/follow', data=payload) as resp:
    return await resp.text()


async def upload_register(session, addr, user):
  payload = {'first_name': 'first_name_' + user, 'last_name': 'last_name_' + user,
             'username': 'username_' + user, 'password': 'password_' + user, 'user_id': user}
  async with session.post(addr + '/wrk2-api/user/register', data=payload) as resp:
    return await resp.text()


async def upload_compose(session, addr, user_id, num_users):
  text = ''.join(random.choices(string.ascii_letters + string.digits, k=256))
  # user mentions
  for _ in range(random.randint(0, 5)):
    text += ' @username_' + str(random.randint(0, num_users))
  # urls
  for _ in range(random.randint(0, 5)):
    text += ' http://' + \
        ''.join(random.choices(string.ascii_lowercase + string.digits, k=64))
  # media
  media_ids = []
  media_types = []
  for _ in range(random.randint(0, 5)):
    media_ids.append('\"' + ''.join(random.choices(string.digits, k=18)) + '\"')
    media_types.append('\"png\"')
  payload = {'username': 'username_' + str(user_id),
             'user_id': str(user_id),
             'text': text,
             'media_ids': '[' + ','.join(media_ids) + ']',
             'media_types': '[' + ','.join(media_types) + ']',
             'post_type': '0'}
  async with session.post(addr + '/wrk2-api/post/compose', data=payload) as resp:
    return await resp.text()


def getNumNodes(file):
  return int(file.readline())


def getEdges(file):
  edges = []
  lines = file.readlines()
  for line in lines:
    edges.append(line.split())
  return edges


def printResults(results):
  result_type_count = {}
  for result in results:
    try:
      result_type_count[result] += 1
    except KeyError:
      result_type_count[result] = 1
  for result_type, count in result_type_count.items():
    if result_type == '' or result_type.startswith('Success'):
      print('Succeeded:', count)
    elif '500 Internal Server Error' in result_type:
      print('Failed:', count, 'Error:', 'Internal Server Error')
    else:
      print('Failed:', count, 'Error:', result_type.strip())


async def register(addr, nodes, limit=200):
  tasks = []
  conn = aiohttp.TCPConnector(limit=limit)
  async with aiohttp.ClientSession(connector=conn) as session:
    print('Registering Users...')
    for i in range(nodes):
      task = asyncio.ensure_future(upload_register(session, addr, str(i)))
      tasks.append(task)
      if i % limit == 0:
        _ = await asyncio.gather(*tasks)
        print(i)
    results = await asyncio.gather(*tasks)
    printResults(results)


async def follow(addr, edges, limit=200):
  idx = 0
  tasks = []
  conn = aiohttp.TCPConnector(limit=limit)
  async with aiohttp.ClientSession(connector=conn) as session:
    print('Adding follows...')
    for edge in edges:
      task = asyncio.ensure_future(
          upload_follow(session, addr, edge[0], edge[1]))
      tasks.append(task)
      task = asyncio.ensure_future(
          upload_follow(session, addr, edge[1], edge[0]))
      tasks.append(task)
      idx += 1
      if idx % limit == 0:
        _ = await asyncio.gather(*tasks)
        print(idx)
    results = await asyncio.gather(*tasks)
    printResults(results)


async def compose(addr, nodes, limit=200):
  idx = 0
  tasks = []
  conn = aiohttp.TCPConnector(limit=limit)
  async with aiohttp.ClientSession(connector=conn) as session:
    print('Composing posts...')
    for i in range(nodes):
      for _ in range(random.randint(0, 20)):  # up to 20 posts per user, average 10
        task = asyncio.ensure_future(upload_compose(session, addr, i+1, nodes))
        tasks.append(task)
        idx += 1
        if idx % limit == 0:
          _ = await asyncio.gather(*tasks)
          print(idx)
    results = await asyncio.gather(*tasks)
    printResults(results)


if __name__ == '__main__':

  parser = argparse.ArgumentParser('DeathStarBench social graph initializer.')
  parser.add_argument(
      '--graph', help='Graph name. (`socfb-Reed98`, `ego-twitter`, or `soc-twitter-follows-mun`)', default='socfb-Reed98')
  parser.add_argument(
      '--ip', help='IP address of socialNetwork NGINX web server. ', default='127.0.0.1')
  parser.add_argument(
      '--port', help='IP port of socialNetwork NGINX web server.', default=8080)
  parser.add_argument('--compose', action='store_true',
                      help='intialize with up to 20 posts per user', default=False)
  parser.add_argument('--limit', type=int, help='total number simultaneous connections', default=200)
  args = parser.parse_args()

  with open(os.path.join('datasets/social-graph', args.graph, f'{args.graph}.nodes'), 'r') as f:
    nodes = getNumNodes(f)
  with open(os.path.join('datasets/social-graph', args.graph, f'{args.graph}.edges'), 'r') as f:
    edges = getEdges(f)

  random.seed(1)   # deterministic random numbers

  addr = 'http://{}:{}'.format(args.ip, args.port)
  limit = args.limit
  loop = asyncio.new_event_loop()
  future = asyncio.ensure_future(register(addr, nodes, limit), loop=loop)
  loop.run_until_complete(future)
  future = asyncio.ensure_future(follow(addr, edges, limit), loop=loop)
  loop.run_until_complete(future)
  if args.compose:
    future = asyncio.ensure_future(compose(addr, nodes, limit), loop=loop)
    loop.run_until_complete(future)
