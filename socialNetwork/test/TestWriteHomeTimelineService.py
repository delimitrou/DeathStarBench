#!/usr/bin/env python
import pika
import json

credentials = pika.PlainCredentials('guest', 'guest')
connection = pika.BlockingConnection(
  pika.ConnectionParameters(host='ath-8.ece.cornell.edu', credentials=credentials))
channel = connection.channel()

channel.queue_declare(queue='write-home-timeline', durable=True)


msg_json = {
  "req_id": 1,
  "post_id": 1,
  "user_id": 1,
  "timestamp": 1,
  "user_mentions_id": [0,2,3],
  "carrier": ""
}

msg = json.dumps(msg_json)

channel.basic_publish(exchange='', routing_key='write-home-timeline', body=msg)
print(" [x] Sent 'Hello World!'")
connection.close()