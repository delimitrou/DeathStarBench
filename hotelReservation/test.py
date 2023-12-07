import requests
import random
import time


def reviews():
    url = "http://localhost:5000/review"
    payload = {"hotelId":"1", "username": "Cornell_0", "password": "0000000000"}
    t_before = time.time()
    r = requests.get(url, params=payload)
    t_after = time.time()
    t = t_after - t_before
    print("Response = ", r.text)
    print("Time = ",t)

reviews()
