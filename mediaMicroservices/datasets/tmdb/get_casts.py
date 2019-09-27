import requests
import json
import optparse
import time

def worker(api_key, movies):
  language="en-US"
  casts = []
  cast_ids = set()
  for movie in movies:
    for cast in movie["cast"]:
      cast_ids.add(cast["id"])
  print("num_of_casts:", len(cast_ids))
  for cast_id in cast_ids:
    cast_url = "https://api.themoviedb.org/3/person/" + str(cast_id)
    r = requests.request("GET", cast_url, params={"language": language, "api_key": api_key})
    if (r.status_code != 200):
      print("Failed to get popular_movie", "status_code:", r.status_code, "message:", r.text)
    casts.append(r.json())
    # rate limit of 4 reqs per second
    time.sleep(0.25)
    print("cast", cast_id, "success")
  return casts

def main():
  parser = optparse.OptionParser()
  parser.add_option("--rfile", type="string", dest="rfile")
  parser.add_option("--wfile", type="string", dest="wfile")
  parser.add_option("--key", type="string", dest="api_key")
  (options, args) = parser.parse_args()
  with open(options.rfile, "r") as movie_file:
    movies = json.load(movie_file)
    casts = worker(options.api_key, movies)
    with open(options.wfile, "w") as cast_file:
      json.dump(casts, cast_file, indent=2)

if __name__ == '__main__':
  main()