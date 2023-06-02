import requests
import time

getfollow_url = 'http://localhost:31992/v1.0/invoke/dapr-social-graph/method/getfollow'
getfollower_url = 'http://localhost:31992/v1.0/invoke/dapr-social-graph/method/getfollower'
follow_url = 'http://localhost:31992/v1.0/invoke/dapr-social-graph/method/follow'
unfollow_url = 'http://localhost:31992/v1.0/invoke/dapr-social-graph/method/unfollow'

# user: follow
# recmd-user-x: recmd-user-0 to recmd-user-9, and recmd-user-3(x+10) to recmd-user-3(x+19)

def make_user(i):
    return 'recmd-user-' + str(i)

for i in range(0, 20):
    u = make_user(i)
    print('------ recmd-user-%d follows users recmd-user-1 to recmd-user-9 ------' %i)
    for j in range(1, 10):
        payload = {
            'send_unix_ms': int(time.time() * 1000),
            'user_id': u,
            'follow_id': make_user(j),
        }
        r = requests.post(follow_url, json=payload)
        print(r.text)
    
    print(
        '------ recmd-user-%d follows users recmd-user-%d to recmd-user-%d, by user id step 3 ------' %(
        i, 3*(i + 10), 3*(i + 19)))
    for j in range(10, 20):
        payload = {
            'send_unix_ms': int(time.time() * 1000),
            'user_id': u,
            'follow_id': make_user(3*(i+j)),
        }
        r = requests.post(follow_url, json=payload)
        print(r.text)

print('## ------------------------------ ##')
for i in range(0, 20):
    print("------ follow list of recmd-user-%d ------" %i)
    payload = {
        'send_unix_ms': int(time.time() * 1000),
        'user_ids': [make_user(i)]
    }
    r = requests.get(getfollow_url, json=payload)
    print(r.text)
