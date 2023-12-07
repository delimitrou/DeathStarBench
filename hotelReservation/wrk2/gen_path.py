import sys;
import os;

of = open("paths.txt", "w")

n = 1000;

n = int(sys.argv[1]);

for i in range(n):
    of.write("/%d.html\n" % (i))

of.close()
