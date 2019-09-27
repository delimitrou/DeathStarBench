import json
import io

with open("movies.json", "r") as in_file:
  movies = json.load(in_file)

  with io.open("movie_titles.txt", "w", encoding='utf8') as out_file:
    for movie in movies:
      try:
        out_file.write("\"" + movie["title"] + "\",\n")
      except:
        print(movie["title"])
