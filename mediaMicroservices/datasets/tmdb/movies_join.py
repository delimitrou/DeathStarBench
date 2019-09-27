import json

file1 =  open("movies_1_10.json", "r")
file2 =  open("movies_11_20.json", "r")
file3 =  open("movies_21_30.json", "r")
file4 =  open("movies_31_40.json", "r")
file5 =  open("movies_41_50.json", "r")

movies = list()
movies += json.load(file1)
movies += json.load(file2)
movies += json.load(file3)
movies += json.load(file4)
movies += json.load(file5)

with open("movies.json", "w") as file:
  json.dump(movies, file, indent=2)