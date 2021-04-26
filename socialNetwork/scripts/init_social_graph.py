import aiohttp
import asyncio
import sys
import string
import random
import argparse

async def upload_follow(session, addr, user_0, user_1):
  payload = {'user_name': 'username_' + user_0, 'followee_name': 'username_' + user_1}
  async with session.post(addr + "/wrk2-api/user/follow", data=payload) as resp:
    return await resp.text()

async def upload_register(session, addr, user):
  payload = {'first_name': 'first_name_' + user, 'last_name': 'last_name_' + user,
             'username': 'username_' + user, 'password': 'password_' + user, 'user_id': user}
  async with session.post(addr + "/wrk2-api/user/register", data=payload) as resp:
    return await resp.text()

async def upload_compose(session, addr, user_id, num_users):
  text = ''.join(random.choices(string.ascii_letters + string.digits, k=256))
  # user mentions
  for _ in range(random.randint(0, 5)):
    text += " @username_" + str(random.randint(0, num_users))
  # urls
  for _ in range(random.randint(0, 5)):
    text += " http://" + ''.join(random.choices(string.ascii_lowercase + string.digits, k=64))
  # media
  media_ids = []
  media_types = []
  for _ in range(random.randint(0, 5)):
    media_ids.append('"' + ''.join(random.choices(string.digits, k=18)) + '"')
    media_types.append('"png"')
  payload = {'username': 'username_' + str(user_id),
             'user_id': str(user_id),
             'text': text,
             'media_ids': "[" + ','.join(media_ids) + "]",
             'media_types': "[" + ','.join(media_types) + "]",
             'post_type': '0'}
  async with session.post(addr + "/wrk2-api/post/compose", data=payload) as resp:
    return await resp.text()

def getNodes(file):
  line = file.readline()
  word = line.split()[0]
  return int(word)

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
    if result_type == '' or result_type.startswith("Success"):
      print("Succeeded:", count)
    elif "500 Internal Server Error" in result_type:
      print("Failed:", count, "Error:", "Internal Server Error")
    else:
      print("Failed:", count, "Error:", result_type.strip())

async def register(addr, nodes):
  idx = 0
  tasks = []
  conn = aiohttp.TCPConnector(limit=200)
  async with aiohttp.ClientSession(connector=conn) as session:
    print("Registering Users...")
    for i in range(1, nodes + 1):
      task = asyncio.ensure_future(upload_register(session, addr, str(i)))
      tasks.append(task)
      idx += 1
      if idx % 200 == 0:
        _ = await asyncio.gather(*tasks)
        print(idx)
    results = await asyncio.gather(*tasks)
    printResults(results)


async def follow(addr, edges):
  idx = 0
  tasks = []
  conn = aiohttp.TCPConnector(limit=200)
  async with aiohttp.ClientSession(connector=conn) as session:
    print("Adding follows...")
    for edge in edges:
      task = asyncio.ensure_future(upload_follow(session, addr, edge[0], edge[1]))
      tasks.append(task)
      task = asyncio.ensure_future(upload_follow(session, addr, edge[1], edge[0]))
      tasks.append(task)
      idx += 1
      if idx % 200 == 0:
        _ = await asyncio.gather(*tasks)
        print(idx)
    results = await asyncio.gather(*tasks)
    printResults(results)


async def compose(addr, nodes):
  idx = 0
  tasks = []
  conn = aiohttp.TCPConnector(limit=200)
  async with aiohttp.ClientSession(connector=conn) as session:
    print("Composing posts...")
    for i in range(nodes):
      for _ in range(random.randint(0,20)): # up to 20 posts per user, average 10
        task = asyncio.ensure_future(upload_compose(session, addr, i+1, nodes))
        tasks.append(task)
        idx += 1
        if idx % 200 == 0:
          _ = await asyncio.gather(*tasks)
          print(idx)
    results = await asyncio.gather(*tasks)
    printResults(results)


if __name__ == '__main__':
  filename_default = "datasets/social-graph/socfb-Reed98/socfb-Reed98.mtx"
  ip_default = "127.0.0.1"
  port_default = "8080"

  parser = argparse.ArgumentParser("DeathStarBench social graph initializer.")
  parser.add_argument("--graph", help="Path to graph file.", default=filename_default)
  parser.add_argument("--ip", help="IP address of socialNetwork NGINX web server.", default=ip_default)
  parser.add_argument("--port", help="IP port of socialNetwork NGINX web server.", default=port_default)
  parser.add_argument("--compose", action="store_true", help="intialize with up to 20 posts per user", default=False)
  args = parser.parse_args()

  with open(args.graph, 'r') as f:
    nodes = getNodes(f)
    edges = getEdges(f)

  random.seed(1)   # deterministic random numbers

  addr = "http://{}:{}".format(args.ip, args.port)
  loop = asyncio.get_event_loop()
  future = asyncio.ensure_future(register(addr, nodes))
  loop.run_until_complete(future)
  future = asyncio.ensure_future(follow(addr, edges))
  loop.run_until_complete(future)
  if args.compose:
    future = asyncio.ensure_future(compose(addr, nodes))
    loop.run_until_complete(future)
