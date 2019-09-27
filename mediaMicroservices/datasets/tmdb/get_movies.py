import requests
import json
import optparse
import time

def worker(api_key, page_start, page_stop):
  movies_url = "https://api.themoviedb.org/3/movie/popular"
  language="en-US"
  movies = []
  for page in range(page_start, page_stop):
    parameter = {"language": language, "page": page, "api_key": api_key}
    r = requests.request("GET", movies_url, params=parameter)
    if (r.status_code != 200):
      print("Failed to get popular_movie", "status_code:", r.status_code, "message:", r.text)
    movies += r.json()["results"]
    for i in range(len(movies)):
      movie_id = movies[i]["id"]
      casts_url = "https://api.themoviedb.org/3/movie/" + str(movie_id) + "/credits"
      r = requests.request("GET", casts_url, params={"api_key": api_key})
      if (r.status_code != 200):
        print("Failed to get popular_movie", "status_code:", r.status_code, "message:", r.text)
      movies[i]["cast"] = r.json()["cast"]
      if (len(movies[i]["cast"]) > 10):
        movies[i]["cast"] = movies[i]["cast"][:10]
      # rate limit of 4 reqs per second
      time.sleep(0.25)
    print("page", page, "success")

  return movies

def main():
  parser = optparse.OptionParser()
  parser.add_option("--start", type="int", dest="start")
  parser.add_option("--stop", type="int", dest="stop")
  parser.add_option("--key", type="string", dest="api_key")
  (options, args) = parser.parse_args()
  movies = worker(options.api_key, options.start, options.stop + 1)
  filename = "movies_" + str(options.start) + "_" + str(options.stop) + ".json"
  with open(filename, "w") as file:
    json.dump(movies, file, indent=2)

if __name__ == '__main__':
  main()



