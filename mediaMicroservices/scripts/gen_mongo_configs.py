import json
import argparse
import os
import errno

parser = argparse.ArgumentParser()

parser.add_argument("-d", "--dir", action="store", dest="dir_path", type=str, required=True)
parser.add_argument("-n", "--name", action="store", nargs='+', dest="service_name", type=str, required=True)
parser.add_argument("-c", "--config", action="store", dest="config_num", type=int, default=1)
parser.add_argument("-r", "--router", action="store", dest="router_num", type=int, default=1)
parser.add_argument("-s", "--shard", action="store", dest="shard_num", type=int, default=3)
parser.add_argument("-p", "--replica", action="store", dest="replica_num", type=int, default=1)

args = parser.parse_args()

for service_name in args.service_name:
  if args.dir_path[-1] != '/':
    path = args.dir_path + "/" + service_name
  else:
    path = args.dir_path + service_name
  try:
    os.mkdir(path)
  except OSError as e:
    if e.errno == errno.EEXIST and os.path.isdir(path):
      pass
    else:
      print ("Creation of the directory %s failed" % service_name)
  
  with open(path + "/init-config.js", "w") as file:
    obj = dict()
    obj["_id"] = service_name + "-mongodb-config"
    obj["configsvr"] = True
    obj["version"] = 1
    obj["members"] = list()
    if args.config_num == 1:
      child_obj = dict()
      child_obj["_id"] = 1
      child_obj["host"] = service_name + "-mongodb-config:27017"
      obj["members"].append(child_obj)
    else:
      for i in range(1, args.config_num + 1):
        child_obj = dict()
        child_obj["_id"] = i
        child_obj["host"] = service_name + "-mongodb-config_" + str(i) + ":27017"
        obj["members"].append(child_obj)
    file.write("rs.initiate(\n")
    json.dump(obj, file, indent=2)
    file.write("\n)")
  
  with open(path + "/init-router.js", "w") as file:
    for i in range(1, args.shard_num + 1):
      for j in range(1, args.replica_num + 1):
        shard_name = service_name + "-mongodb-shard-" + str(i)
        if args.replica_num == 1:
          hostname = service_name + "-mongodb-shard-" + str(i)
        else:
          hostname = service_name + "-mongodb-shard-" + str(i) + "_" + str(j)
        file.write("sh.addShard(\"" + shard_name + "/" + hostname + ":27017\")\n")
      file.write("\n")
  
  for i in range(1, args.shard_num + 1):
    with open(path + "/init-shard_" + str(i) + ".js", "w") as file:
      obj = dict()
      obj["_id"] = service_name + "-mongodb-shard-" + str(i)
      obj["version"] = 1
      obj["members"] = list()
      if args.replica_num == 0:
        child_obj = dict()
        child_obj["_id"] = 1
        child_obj["host"] = service_name + "-mongodb-shard-" + str(i) + ":27017"
        obj["members"].append(child_obj)
      else:
        for j in range(1, args.replica_num + 1):
          child_obj = dict()
          child_obj["_id"] = j
          child_obj["host"] = service_name + "-mongodb-shard-" + str(i) + "_" + str(j) + ":27017"
          obj["members"].append(child_obj)
      file.write("rs.initiate(\n")
      json.dump(obj, file, indent=2)
      file.write("\n)")
  
  
