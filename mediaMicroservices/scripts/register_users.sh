#!/usr/bin/env bash

for i in {1..1000}; do
  curl -d "first_name=first_name_"$i"&last_name=last_name_"$i"&username=username_"$i"&password=password_"$i \
      http://10.68.131.177:8080/wrk2-api/user/register
  # curl -d "first_name=first_name_"$i"&last_name=last_name_"$i"&username=username_"$i"&password=password_"$i \
  #     http://10.0.0.6:8082/wrk2-api/user/register
done