import requests
import time

getfollow_url = 'http://localhost:31992/v1.0/invoke/dapr-social-graph/method/getfollow'
getfollower_url = 'http://localhost:31992/v1.0/invoke/dapr-social-graph/method/getfollower'
follow_url = 'http://localhost:31992/v1.0/invoke/dapr-social-graph/method/follow'
unfollow_url = 'http://localhost:31992/v1.0/invoke/dapr-social-graph/method/unfollow'

print("getfollow non-existent users")
payload = {
    'send_unix_ms': int(time.time() * 1000),
    'user_ids': ['2233', 'bilibili', 'baiNianJi']
}
r = requests.get(getfollow_url, json=payload)
print(r.text)