import requests
import time

getfollow_url = 'http://localhost:31992/v1.0/invoke/dapr-social-graph/method/getfollow'
getfollower_url = 'http://localhost:31992/v1.0/invoke/dapr-social-graph/method/getfollower'
follow_url = 'http://localhost:31992/v1.0/invoke/dapr-social-graph/method/follow'
unfollow_url = 'http://localhost:31992/v1.0/invoke/dapr-social-graph/method/unfollow'

# assume starting from an empty store
#---------------- Test: follow non-existing relation ----------------#
print("------ follow non-existent users ------")
payload = {
    'send_unix_ms': int(time.time() * 1000),
    'user_id': '2233',
    'follow_id': 'bilibili',
}
r = requests.post(follow_url, json=payload)
print(r.text)

print("------ Check follow list of 2233 (bilibili) ------")
payload = {
    'send_unix_ms': int(time.time() * 1000),
    'user_ids': ['2233']
}
r = requests.get(getfollow_url, json=payload)
print(r.text)

print("------ Check follower list of bilibili (2233) ------")
payload = {
    'send_unix_ms': int(time.time() * 1000),
    'user_ids': ['bilibili']
}
r = requests.get(getfollower_url, json=payload)
print(r.text)


#---------------- Test: repetitive follow ----------------#
print("\n------  follow repetitive users ------")
payload = {
    'send_unix_ms': int(time.time() * 1000),
    'user_id': '2233',
    'follow_id': 'bilibili',
}
r = requests.post(follow_url, json=payload)
print(r.text)

print("------ Check follow list of 2233 (bilibili) ------")
payload = {
    'send_unix_ms': int(time.time() * 1000),
    'user_ids': ['2233']
}
r = requests.get(getfollow_url, json=payload)
print(r.text)

print("------ Check follower list of bilibili (2233) ------")
payload = {
    'send_unix_ms': int(time.time() * 1000),
    'user_ids': ['bilibili']
}
r = requests.get(getfollower_url, json=payload)
print(r.text)

#---------------- Test: more follow ----------------#
print("\n------  follow of another user ------")
payload = {
    'send_unix_ms': int(time.time() * 1000),
    'user_id': 'BaiNianJi',
    'follow_id': 'bilibili',
}
r = requests.post(follow_url, json=payload)
print(r.text)

payload = {
    'send_unix_ms': int(time.time() * 1000),
    'user_id': 'BaiNianJi',
    'follow_id': '2233',
}
r = requests.post(follow_url, json=payload)
print(r.text)

print("------ Check follow list of BaiNianJi (2233, bilibili) ------")
payload = {
    'send_unix_ms': int(time.time() * 1000),
    'user_ids': ['BaiNianJi']
}
r = requests.get(getfollow_url, json=payload)
print(r.text)

print("------ Check follower list of 2233 (BaiNianJi) ------")
payload = {
    'send_unix_ms': int(time.time() * 1000),
    'user_ids': ['2233']
}
r = requests.get(getfollower_url, json=payload)
print(r.text)

print("------ Check follower list of bilibili (2233, BaiNianJi) ------")
payload = {
    'send_unix_ms': int(time.time() * 1000),
    'user_ids': ['bilibili']
}
r = requests.get(getfollower_url, json=payload)
print(r.text)

#---------------- Test: unfollow non-existing user ----------------#
print("\n------  unfollow non-existing user ------")
payload = {
    'send_unix_ms': int(time.time() * 1000),
    'user_id': '2233',
    'unfollow_id': 'AcNiang',
}
r = requests.post(unfollow_url, json=payload)
print(r.text)
print("------ Check follower list of 2233 (BaiNianJi) ------")
payload = {
    'send_unix_ms': int(time.time() * 1000),
    'user_ids': ['2233']
}
r = requests.get(getfollower_url, json=payload)
print(r.text)

#---------------- Test: unfollow existing user ----------------#
print("\n------  unfollow existing user ------")
payload = {
    'send_unix_ms': int(time.time() * 1000),
    'user_id': 'BaiNianJi',
    'unfollow_id': 'bilibili',
}
r = requests.post(unfollow_url, json=payload)
print(r.text)

print("------ Check follower list of bilibili (BaiNianJi) ------")
payload = {
    'send_unix_ms': int(time.time() * 1000),
    'user_ids': ['bilibili']
}
r = requests.get(getfollower_url, json=payload)
print(r.text)

print("------ Check follow list of BaiNianJi (2233) ------")
payload = {
    'send_unix_ms': int(time.time() * 1000),
    'user_ids': ['BaiNianJi']
}
r = requests.get(getfollow_url, json=payload)
print(r.text)