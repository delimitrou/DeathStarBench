# -*- coding: UTF-8 -*-

import json
import os
import yaml

def config_nginx(tls):
    if tls:
        f = open('/usr/local/openresty/nginx/conf/nginx.conf')
        content = f.read()

        content = content.replace('config:set("ssl", false)', 'config:set("ssl", true)')

        f.close()
        f = open('/usr/local/openresty/nginx/conf/nginx.conf', 'w')
        f.write(content)
        f.close()
    else:
        f = open('/usr/local/openresty/nginx/conf/nginx.conf')
        content = f.read()

        content = content.replace('config:set("ssl", true)', 'config:set("ssl", false)')

        f.close()

        f = open('/usr/local/openresty/nginx/conf/nginx.conf', 'w')
        f.write(content)
        f.close()

def config_thrift(tls):
    if tls:
        f = open('/social-network-microservices/config/service-config.json')
        content = f.read()
        j = json.loads(content)

        j['ssl']['enabled'] = True

        f.close()
        f = open('/social-network-microservices/config/service-config.json', 'w')
        f.write(json.dumps(j, indent=2))
        f.close()
    else:
        f = open('/social-network-microservices/config/service-config.json')
        content = f.read()
        j = json.loads(content)

        j['ssl']['enabled'] = False

        f.close()
        f = open('/social-network-microservices/config/service-config.json', 'w')
        f.write(json.dumps(j, indent=2))
        f.close()

def config_mongod(tls):
    if tls:
        f = open('/social-network-microservices/config/mongod.conf')
        content = f.read()
        y = yaml.load(content)

        y['net']['tls']['mode'] = 'requireTLS'
        y['net']['tls']['certificateKeyFile'] = '/keys/server.pem'

        f.close()
        f = open('/social-network-microservices/config/mongod.conf', 'w')
        f.write(yaml.dump(y, default_flow_style=False))
        f.close()
    else:
        f = open('/social-network-microservices/config/mongod.conf')
        content = f.read()
        y = yaml.load(content)

        y['net']['tls']['mode'] = 'disabled'
        try:
            del y['net']['tls']['certificateKeyFile']
        except:
            pass

        f.close()
        f = open('/social-network-microservices/config/mongod.conf', 'w')
        f.write(yaml.dump(y, default_flow_style=False))
        f.close()

def config_redis(tls):
    if tls:
        f = open('/social-network-microservices/config/redis.conf')
        content = f.read()

        content = content.replace('port 6379', 'port 0')
        content = content.replace('tls-port 0', 'tls-port 6379')

        f.close()
        f = open('/social-network-microservices/config/redis.conf', 'w')
        f.write(content)
        f.close()
    else:
        f = open('/social-network-microservices/config/redis.conf')
        content = f.read()

        content = content.replace('port 0', 'port 6379')
        content = content.replace('tls-port 6379', 'tls-port 0')

        f.close()
        f = open('/social-network-microservices/config/redis.conf', 'w')
        f.write(content)
        f.close()

tls = True
tls_str = os.environ.get('TLS', '0').lower()
if tls_str == '0' or tls_str == 'false':
    tls = False

config_nginx(tls)
config_thrift(tls)
config_mongod(tls)
config_redis(tls)
