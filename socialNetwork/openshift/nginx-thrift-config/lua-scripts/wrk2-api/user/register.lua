local _M = {}

local function _StrIsEmpty(s)
  return s == nil or s == ''
end

local function _NgxInternalError(ngx, label, msg)
  ngx.status = ngx.HTTP_INTERNAL_SERVER_ERROR
  local ErrorMessage = "<no message>"
  if not _StrIsEmpty(msg) then
    ErrorMessage = msg
  end
  ngx.say(label .. ErrorMessage)
  ngx.log(ngx.ERR, label .. ErrorMessage)
  ngx.exit(ngx.HTTP_INTERNAL_SERVER_ERROR)
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
      _StrIsEmpty(post.username) or _StrIsEmpty(post.password) or
      _StrIsEmpty(post.user_id)) then
    ngx.status = ngx.HTTP_BAD_REQUEST
    ngx.say("Incomplete arguments")
    ngx.log(ngx.ERR, "Incomplete arguments")
    ngx.exit(ngx.HTTP_BAD_REQUEST)
  end

  local client = GenericObjectPool:connection(UserServiceClient, "user-service.social-network.svc.cluster.local", 9090)

  local status, err = pcall(client.RegisterUserWithId, client, req_id, post.first_name,
      post.last_name, post.username, post.password, tonumber(post.user_id), carrier)
  GenericObjectPool:returnConnection(client)

  if not status then
    _NgxInternalError(ngx, "User registration failure: ", err.message)
  else
    ngx.say("Successfully registered.")
  end
  span:finish()
end

return _M