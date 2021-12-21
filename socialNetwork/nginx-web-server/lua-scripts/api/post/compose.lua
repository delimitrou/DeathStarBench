local _M = {}
local k8s_suffix = os.getenv("fqdn_suffix")
if (k8s_suffix == nil) then
  k8s_suffix = ""
end

local function _StrIsEmpty(s)
  return s == nil or s == ''
end

function _M.ComposePost()

  local bridge_tracer = require "opentracing_bridge_tracer"
  local ngx = ngx
  local cjson = require "cjson"
  local jwt = require "resty.jwt"

  local GenericObjectPool = require "GenericObjectPool"
  local social_network_ComposePostService = require "social_network_ComposePostService"
  local ComposePostServiceClient = social_network_ComposePostService.ComposePostServiceClient

  GenericObjectPool:setMaxTotal(512)

  local req_id = tonumber(string.sub(ngx.var.request_id, 0, 15), 16)
  local tracer = bridge_tracer.new_from_global()
  local parent_span_context = tracer:binary_extract(ngx.var.opentracing_binary_context)

  ngx.req.read_body()
  local post = ngx.req.get_post_args()

  if (_StrIsEmpty(post.post_type) or _StrIsEmpty(post.text)) then
    ngx.status = ngx.HTTP_BAD_REQUEST
    ngx.say("Incomplete arguments")
    ngx.log(ngx.ERR, "Incomplete arguments")
    ngx.exit(ngx.HTTP_BAD_REQUEST)
  end

  if (_StrIsEmpty(ngx.var.cookie_login_token)) then
    ngx.status = ngx.HTTP_UNAUTHORIZED
    ngx.redirect("../../index.html")
    ngx.exit(ngx.HTTP_OK)
  end

  local login_obj = jwt:verify(ngx.shared.config:get("secret"), ngx.var.cookie_login_token)
  if not login_obj["verified"] then
    ngx.status = ngx.HTTP_UNAUTHORIZED
    ngx.say(login_obj.reason);
    ngx.redirect("../../index.html")
    ngx.exit(ngx.HTTP_OK)
  end
  -- get user id/name from login obj
  local timestamp = tonumber(login_obj["payload"]["timestamp"])
  local ttl = tonumber(login_obj["payload"]["ttl"])
  local user_id = tonumber(login_obj["payload"]["user_id"])
  local username = login_obj["payload"]["username"]

  if (timestamp + ttl < ngx.time()) then
    ngx.status = ngx.HTTP_UNAUTHORIZED
    ngx.header.content_type = "text/plain"
    ngx.say("Login token expired, please log in again")
    ngx.redirect("../../index.html")
    ngx.exit(ngx.HTTP_OK)
  else
    local status, ret
    local client = GenericObjectPool:connection(
      ComposePostServiceClient, "compose-post-service" .. k8s_suffix, 9090)

    local span = tracer:start_span("compose_post_client",
      { ["references"] = { { "child_of", parent_span_context } } })
    local carrier = {}
    tracer:text_map_inject(span:context(), carrier)

    if (not _StrIsEmpty(post.media_ids) and not _StrIsEmpty(post.media_types)) then
      status, ret = pcall(client.ComposePost, client,
          req_id, username, tonumber(user_id), post.text,
          cjson.decode(post.media_ids), cjson.decode(post.media_types),
          tonumber(post.post_type), carrier)
    else
      status, ret = pcall(client.ComposePost, client,
          req_id, username, tonumber(user_id), post.text,
          {}, {}, tonumber(post.post_type), carrier)
    end

    if not status then
      ngx.status = ngx.HTTP_INTERNAL_SERVER_ERROR
      if (ret.message) then
        ngx.say("compost_post failure: " .. ret.message)
        ngx.log(ngx.ERR, "compost_post failure: " .. ret.message)
      else
        ngx.say("compost_post failure: " .. ret)
        ngx.log(ngx.ERR, "compost_post failure: " .. ret)
      end
      client.iprot.trans:close()
      ngx.exit(ngx.status)
    end

    GenericObjectPool:returnConnection(client)
    ngx.status = ngx.HTTP_OK
    ngx.say("Successfully upload post")
    span:finish()
    ngx.exit(ngx.status)
  end
end

return _M