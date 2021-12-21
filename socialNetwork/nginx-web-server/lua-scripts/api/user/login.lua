local _M = {}
local k8s_suffix = os.getenv("fqdn_suffix")
if (k8s_suffix == nil) then
  k8s_suffix = ""
end

local function _StrIsEmpty(s)
  return s == nil or s == ''
end

function _M.Login()
  local bridge_tracer = require "opentracing_bridge_tracer"
  local ngx = ngx
  local GenericObjectPool = require "GenericObjectPool"
  local UserServiceClient = require "social_network_UserService".UserServiceClient
  local cjson = require "cjson"

  local req_id = tonumber(string.sub(ngx.var.request_id, 0, 15), 16)
  local tracer = bridge_tracer.new_from_global()
  local parent_span_context = tracer:binary_extract(
      ngx.var.opentracing_binary_context)
  local span = tracer:start_span("Login",
      {["references"] = {{"child_of", parent_span_context}}})
  local carrier = {}
  tracer:text_map_inject(span:context(), carrier)

  ngx.req.read_body()
  local args = ngx.req.get_post_args()

  if (_StrIsEmpty(args.username) or _StrIsEmpty(args.password)) then
    ngx.status = ngx.HTTP_BAD_REQUEST
    ngx.header.content_type = "text/plain"
    ngx.say("Incomplete arguments")
    ngx.say(ngx.var.scheme .. "://" .. ngx.var.server_addr .. ":" .. ngx.var.server_port)
    ngx.log(ngx.ERR, "Incomplete arguments")
    ngx.exit(ngx.HTTP_BAD_REQUEST)
    return ngx.redirect("/login.html")
  end

  local client = GenericObjectPool:connection(UserServiceClient, "user-service" .. k8s_suffix, 9090)

  local status, ret = pcall(client.Login, client, req_id,
      args.username, args.password, carrier)
  GenericObjectPool:returnConnection(client)

  if not status then
    ngx.status = ngx.HTTP_INTERNAL_SERVER_ERROR
    if (ret.message) then
      ngx.header.content_type = "text/plain"
      ngx.say("User login failure: " .. ret.message)
      ngx.log(ngx.ERR, "User login failure: " .. ret.message)
    else
      ngx.header.content_type = "text/plain"
      ngx.say("User login failure: " .. ret.message)
      ngx.log(ngx.ERR, "User login failure: " .. ret.message)
    end
    ngx.exit(ngx.HTTP_OK)
  else
    ngx.header.content_type = "text/plain"
    ngx.header["Set-Cookie"] = "login_token=" .. ret .. "; Path=/; Expires="
        .. ngx.cookie_time(ngx.time() + ngx.shared.config:get("cookie_ttl"))

    ngx.redirect("../../main.html?username=" .. args.username)
    ngx.exit(ngx.HTTP_OK)
  end
  span:finish()
end

return _M