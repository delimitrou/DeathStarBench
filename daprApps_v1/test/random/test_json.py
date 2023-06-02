import json
from typing import Dict

def makeMeta(i: str, objects: Dict[str, str]):
    return {
        'post_id': i,
        'objects': objects,
    }

d = [
    {'label': 'raptor', 'score': 0.9, 'box': [{'upper left': 10, 'upper right': 20}]},
    {'label': 'lightning', 'score': 0.7, 'box': [{'upper left': 15, 'upper right': 25}]},
]
dj = json.dumps(d)
objects = {'red-flag': dj}
m = makeMeta('red-flag-shows', objects)
print(m)
mj = json.dumps(m)
print('-----------------')
print(mj)
print('-----------------')
md = json.loads(mj)
for o in md['objects']:
    print(o)
    print(json.loads(md['objects'][o]))
# od = json.loads(md['objects'])
# print(od)