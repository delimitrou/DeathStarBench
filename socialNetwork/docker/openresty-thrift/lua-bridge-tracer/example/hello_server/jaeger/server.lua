local bridge_tracer = require('opentracing_bridge_tracer')
local http = require('http')

local f = assert(io.open('/etc/jaeger-config.json', "rb"))
local config = f:read("*all")
f:close()

local tracer = bridge_tracer:new('/usr/local/lib/libjaegertracing_plugin.so', config)

http.createServer(function (req, res)
  local span = tracer:start_span('hello')
  local body = "Hello world\n"
  res:setHeader("Content-Type", "text/plain")
  res:setHeader("Content-Length", #body)
  res:finish(body)
  span:finish()
end):listen(8080)

print('Server running  on 8080/')
