local _M = {}

local function _StrIsEmpty(s)
  return s == nil or s == ''
end

function _M.RegisterUser()
  local bridge_tracer = require "opentracing_bridge_tracer"
  local ngx = ngx
  local GenericObjectPool = require "GenericObjectPool"
  local UserServiceClient = require "social_network_UserService"

  local req_id = tonumber(string.sub(ngx.var.request_id, 0, 15), 16)
  local tracer = bridge_tracer.new_from_global()
  local parent_span_context = tracer:binary_extract(
      ngx.var.opentracing_binary_context)
  local span = tracer:start_span("RegisterUser",
      {["references"] = {{"child_of", parent_span_context}}})
  local carrier = {}
  tracer:text_map_inject(span:context(), carrier)

  ngx.req.read_body()
  local post = ngx.req.get_post_args()

  if (_StrIsEmpty(post.first_name) or _StrIsEmpty(post.last_name) or
      _StrIsEmpty(post.username) or _StrIsEmpty(post.password)) then
    ngx.status = ngx.HTTP_BAD_REQUEST
    ngx.say("Incomplete arguments")
    ngx.log(ngx.ERR, "Incomplete arguments")
    ngx.exit(ngx.HTTP_BAD_REQUEST)
  end

  local client = GenericObjectPool:connection(UserServiceClient, "user-service.social-network.svc.cluster.local", 9090)

  local status, err = pcall(client.RegisterUser, client, req_id, post.first_name,
      post.last_name, post.username, post.password, carrier)
  GenericObjectPool:returnConnection(client)

  if not status then
    ngx.status = ngx.HTTP_INTERNAL_SERVER_ERROR
    if (err.message) then
      ngx.say("User registration failure: " .. err.message)
      ngx.log(ngx.ERR, "User registration failure: " .. err.message)
    else
      ngx.say("User registration failure: " .. err.message)
      ngx.log(ngx.ERR, "User registration failure: " .. err.message)
    end
    ngx.exit(ngx.HTTP_INTERNAL_SERVER_ERROR)
  end


  span:finish()
end

return _M