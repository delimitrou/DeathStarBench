FROM yg397/openresty-thrift:xenial

COPY ../../../../gen-lua /gen-lua
COPY ../../lua-thrift /usr/local/openresty/lualib/thrift

COPY ../../../../nginx-web-server/lua-scripts /usr/local/openresty/nginx/lua-scripts
COPY ../../../../nginx-web-server/pages /usr/local/openresty/nginx/pages
