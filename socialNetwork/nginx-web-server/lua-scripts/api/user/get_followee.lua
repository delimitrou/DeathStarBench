local _M = {}
local k8s_suffix = os.getenv("fqdn_suffix")
if (k8s_suffix == nil) then
  k8s_suffix = ""
end

local function _StrIsEmpty(s)
  return s == nil or s == ''
end

local function _LoadFollwee(data)
  local followee_list = {}
  for _, followee in ipairs(data) do
    local new_followee = {}
    new_followee["followee_id"] = tostring(followee)
    table.insert(followee_list, new_followee)
  end
  return followee_list
end





function _M.GetFollowee()
  local bridge_tracer = require "opentracing_bridge_tracer"
  local ngx = ngx
  local GenericObjectPool = require "GenericObjectPool"
  local SocialGraphServiceClient = require "social_network_SocialGraphService".SocialGraphServiceClient
  local cjson = require "cjson"
  local jwt = require "resty.jwt"
  local liblualongnumber = require "liblualongnumber"

  local req_id = tonumber(string.sub(ngx.var.request_id, 0, 15), 16)
  local tracer = bridge_tracer.new_from_global()
  local parent_span_context = tracer:binary_extract(
      ngx.var.opentracing_binary_context)
  local span = tracer:start_span("GetFollowee",
      {["references"] = {{"child_of", parent_span_context}}})
  local carrier = {}
  tracer:text_map_inject(span:context(), carrier)

  if (_StrIsEmpty(ngx.var.cookie_login_token)) then
    ngx.status = ngx.HTTP_UNAUTHORIZED
    ngx.exit(ngx.HTTP_OK)
  end

  local login_obj = jwt:verify(ngx.shared.config:get("secret"), ngx.var.cookie_login_token)
  if not login_obj["verified"] then
    ngx.status = ngx.HTTP_UNAUTHORIZED
    ngx.say(login_obj.reason);
    ngx.exit(ngx.HTTP_OK)
  end

  local timestamp = tonumber(login_obj["payload"]["timestamp"])
  local ttl = tonumber(login_obj["payload"]["ttl"])
  local user_id = tonumber(login_obj["payload"]["user_id"])

  if (timestamp + ttl < ngx.time()) then
    ngx.status = ngx.HTTP_UNAUTHORIZED
    ngx.header.content_type = "text/plain"
    ngx.say("Login token expired, please log in again")
    ngx.exit(ngx.HTTP_OK)
  else
    local client = GenericObjectPool:connection(
      SocialGraphServiceClient, "social-graph-service" .. k8s_suffix, 9090)
    local status, ret = pcall(client.GetFollowees, client, req_id,
        user_id, carrier)
    GenericObjectPool:returnConnection(client)
    if not status then
      ngx.status = ngx.HTTP_INTERNAL_SERVER_ERROR
      if (ret.message) then
        ngx.header.content_type = "text/plain"
        ngx.say("Get user-timeline failure: " .. ret.message)
        ngx.log(ngx.ERR, "Get user-timeline failure: " .. ret.message)
      else
        ngx.header.content_type = "text/plain"
        ngx.say("Get user-timeline failure: " .. ret.message)
        ngx.log(ngx.ERR, "Get user-timeline failure: " .. ret.message)
      end
      ngx.exit(ngx.HTTP_INTERNAL_SERVER_ERROR)
    else
      local followee_list = _LoadFollwee(ret)
      ngx.header.content_type = "application/json; charset=utf-8"
      ngx.say(cjson.encode(followee_list) )
    end
  end
end
return _M