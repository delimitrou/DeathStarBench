import json

file1 =  open("casts_1_10.json", "r")
file2 =  open("casts_11_20.json", "r")
file3 =  open("casts_21_30.json", "r")
file4 =  open("casts_31_40.json", "r")
file5 =  open("casts_41_50.json", "r")

casts = list()
casts += json.load(file1)
casts += json.load(file2)
casts += json.load(file3)
casts += json.load(file4)
casts += json.load(file5)

with open("casts.json", "w") as file:
  unique_casts = []
  ids = []
  for cast in casts:
    try:
      if cast["id"] not in ids:
        ids.append(cast["id"])
        unique_casts.append(cast)
    except:
      pass
  json.dump(unique_casts, file, indent=2)
