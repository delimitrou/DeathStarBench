local _M = {}

local dump
dump = function (o)
  if type(o) == 'table' then
    local s = '{ '
    for k,v in pairs(o) do
      if type(k) ~= 'number' then k = '"'..k..'"' end
      s = s .. '['..k..'] = ' .. dump(v) .. ','
    end
    return s .. '} '
  else
    return tostring(o)
  end
end

function _M.test()

  local bridge_tracer = require("opentracing_bridge_tracer")
  local tracer = bridge_tracer.new_from_global()
  local parent_context = tracer:binary_extract(ngx.var.opentracing_binary_context)

  local carrier = {}
  tracer:text_map_inject(parent_context, carrier)
  local parent_context1 = tracer:text_map_extract(carrier)
  local span = tracer:start_span("lua-hello", {["references"] = {{"child_of", parent_context1}}})
  ngx.say("<p>hello, world</p>")
  ngx.say(dump(carrier))
  span:finish()
end

return _M