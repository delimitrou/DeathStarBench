local _M = {}
local k8s_suffix = os.getenv("fqdn_suffix")
if (k8s_suffix == nil) then
  k8s_suffix = ""
end

local function _StrIsEmpty(s)
  return s == nil or s == ''
end

function _M.Follow()
  local bridge_tracer = require "opentracing_bridge_tracer"
  local ngx = ngx
  local GenericObjectPool = require "GenericObjectPool"
  local SocialGraphServiceClient = require "social_network_SocialGraphService".SocialGraphServiceClient
  local jwt = require "resty.jwt"

  local req_id = tonumber(string.sub(ngx.var.request_id, 0, 15), 16)
  local tracer = bridge_tracer.new_from_global()
  local parent_span_context = tracer:binary_extract(
      ngx.var.opentracing_binary_context)
  local span = tracer:start_span("Follow",
      {["references"] = {{"child_of", parent_span_context}}})
  local carrier = {}
  tracer:text_map_inject(span:context(), carrier)

  ngx.req.read_body()
  local post = ngx.req.get_post_args()

  local client = GenericObjectPool:connection(
      SocialGraphServiceClient, "social-graph-service" .. k8s_suffix, 9090)

  -- -- new start --
  -- if (_StrIsEmpty(ngx.var.cookie_login_token)) then
  --   ngx.status = ngx.HTTP_UNAUTHORIZED
  --   ngx.exit(ngx.HTTP_OK)
  -- end

  -- local login_obj = jwt:verify(ngx.shared.config:get("secret"), ngx.var.cookie_login_token)
  -- if not login_obj["verified"] then
  --   ngx.status = ngx.HTTP_UNAUTHORIZED
  --   ngx.say(login_obj.reason);
  --   ngx.exit(ngx.HTTP_OK)
  -- end
  -- -- get user id/name from login obj
  -- local user_id = tonumber(login_obj["payload"]["user_id"])
  -- local username = login_obj["payload"]["username"]

  -- -- new end --


  local status
  local err
  if (not _StrIsEmpty(post.user_id) and not _StrIsEmpty(post.followee_id)) then
    status, err = pcall(client.Follow, client,req_id,
        tonumber(post.user_id), tonumber(post.followee_id), carrier )
  elseif (not _StrIsEmpty(post.user_name) and not _StrIsEmpty(post.followee_name)) then
    status, err = pcall(client.FollowWithUsername, client,req_id,
        post.user_name, post.followee_name, carrier)
  else
    ngx.status = ngx.HTTP_BAD_REQUEST
    ngx.say("Incomplete arguments")
    ngx.log(ngx.ERR, "Incomplete arguments")
    ngx.exit(ngx.HTTP_BAD_REQUEST)
  end

  if not status then
    ngx.status = ngx.HTTP_INTERNAL_SERVER_ERROR
    ngx.say("Follow Failed: " .. err.message)
    ngx.log(ngx.ERR, "Follow Failed: " .. err.message)
    ngx.exit(ngx.HTTP_INTERNAL_SERVER_ERROR)
  end

  GenericObjectPool:returnConnection(client)
  span:finish()
  ngx.redirect("../../contact.html")
  -- ngx.header.content_type = "application/json; charset=utf-8"
  -- ngx.say(cjson.encode(home_timeline) )
  ngx.exit(ngx.HTTP_OK)


end

return _M