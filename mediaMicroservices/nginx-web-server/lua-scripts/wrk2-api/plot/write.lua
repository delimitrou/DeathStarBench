local _M = {}
local k8s_suffix = os.getenv("fqdn_suffix")
if (k8s_suffix == nil) then
  k8s_suffix = ""
end

function _M.WritePlot()
  local bridge_tracer = require "opentracing_bridge_tracer"
  local GenericObjectPool = require "GenericObjectPool"
  local PlotServiceClient = require 'media_service_PlotService'
  local ngx = ngx
  local cjson = require("cjson")

  local req_id = tonumber(string.sub(ngx.var.request_id, 0, 15), 16)
  local tracer = bridge_tracer.new_from_global()
  local parent_span_context = tracer:binary_extract(ngx.var.opentracing_binary_context)
  local span = tracer:start_span("WritePlot  ", {["references"] = {{"child_of", parent_span_context}}})
  local carrier = {}
  tracer:text_map_inject(span:context(), carrier)

  ngx.req.read_body()
  local data = ngx.req.get_body_data()

  if not data then
    ngx.status = ngx.HTTP_BAD_REQUEST
    ngx.say("Empty body")
    ngx.log(ngx.ERR, "Empty body")
    ngx.exit(ngx.HTTP_BAD_REQUEST)
  end

  local plot = cjson.decode(data)
  if (plot["plot_id"] == nil or plot["plot"] == nil) then
    ngx.status = ngx.HTTP_BAD_REQUEST
    ngx.say("Incomplete arguments")
    ngx.log(ngx.ERR, "Incomplete arguments")
    ngx.exit(ngx.HTTP_BAD_REQUEST)
  end

  local client = GenericObjectPool:connection(PlotServiceClient, "plot-service" .. k8s_suffix, 9090)
  client:WritePlot(req_id, plot["plot_id"], plot["plot"], carrier)
  GenericObjectPool:returnConnection(client)
end

return _M