local _M = {}
local k8s_suffix = os.getenv("fqdn_suffix")
if (k8s_suffix == nil) then
  k8s_suffix = ""
end

local function _StrIsEmpty(s)
  return s == nil or s == ''
end

local function _LoadTimeline(data)
  local user_timeline = {}
  for _, timeline_post in ipairs(data) do
    local new_post = {}
    new_post["post_id"] = tostring(timeline_post.post_id)
    new_post["creator"] = {}
    new_post["creator"]["user_id"] = tostring(timeline_post.creator.user_id)
    new_post["creator"]["username"] = timeline_post.creator.username
    new_post["req_id"] = tostring(timeline_post.req_id)
    new_post["text"] = timeline_post.text
    new_post["user_mentions"] = {}
    for _, user_mention in ipairs(timeline_post.user_mentions) do
      local new_user_mention = {}
      new_user_mention["user_id"] = tostring(user_mention.user_id)
      new_user_mention["username"] = user_mention.username
      table.insert(new_post["user_mentions"], new_user_mention)
    end
    new_post["media"] = {}
    for _, media in ipairs(timeline_post.media) do
      local new_media = {}
      new_media["media_id"] = tostring(media.media_id)
      new_media["media_type"] = media.media_type
      table.insert(new_post["media"], new_media)
    end
    new_post["urls"] = {}
    for _, url in ipairs(timeline_post.urls) do
      local new_url = {}
      new_url["shortened_url"] = url.shortened_url
      new_url["expanded_url"] = url.expanded_url
      table.insert(new_post["urls"], new_url)
    end
    new_post["timestamp"] = tostring(timeline_post.timestamp)
    new_post["post_type"] = timeline_post.post_type
    table.insert(user_timeline, new_post)
  end
  return user_timeline
end

function _M.ReadUserTimeline()
  local bridge_tracer = require "opentracing_bridge_tracer"
  local ngx = ngx
  local GenericObjectPool = require "GenericObjectPool"
  local social_network_UserTimelineService = require "social_network_UserTimelineService"
  local UserTimelineServiceClient = social_network_UserTimelineService.UserTimelineServiceClient
  local cjson = require "cjson"
  local liblualongnumber = require "liblualongnumber"

  local req_id = tonumber(string.sub(ngx.var.request_id, 0, 15), 16)
  local tracer = bridge_tracer.new_from_global()
  local parent_span_context = tracer:binary_extract(
      ngx.var.opentracing_binary_context)

  local span = tracer:start_span("ReadUserTimeline",
      {["references"] = {{"child_of", parent_span_context}}})
  local carrier = {}
  tracer:text_map_inject(span:context(), carrier)

  ngx.req.read_body()
  local args = ngx.req.get_uri_args()

  if (_StrIsEmpty(args.user_id) or _StrIsEmpty(args.start) or _StrIsEmpty(args.stop)) then
    ngx.status = ngx.HTTP_BAD_REQUEST
    ngx.say("Incomplete arguments")
    ngx.log(ngx.ERR, "Incomplete arguments")
    ngx.exit(ngx.HTTP_BAD_REQUEST)
  end


  local client = GenericObjectPool:connection(
      UserTimelineServiceClient, "user-timeline-service" .. k8s_suffix, 9090)
  local status, ret = pcall(client.ReadUserTimeline, client, req_id,
      tonumber(args.user_id), tonumber(args.start), tonumber(args.stop), carrier)
  if not status then
    ngx.status = ngx.HTTP_INTERNAL_SERVER_ERROR
    if (ret.message) then
      ngx.say("Get user-timeline failure: " .. ret.message)
      ngx.log(ngx.ERR, "Get user-timeline failure: " .. ret.message)
    else
      ngx.say("Get user-timeline failure: " .. ret)
      ngx.log(ngx.ERR, "Get user-timeline failure: " .. ret)
    end
    client.iprot.trans:close()
    span:finish()
    ngx.exit(ngx.HTTP_INTERNAL_SERVER_ERROR)
  else
    GenericObjectPool:returnConnection(client)
    local user_timeline = _LoadTimeline(ret)
    ngx.header.content_type = "application/json; charset=utf-8"
    ngx.say(cjson.encode(user_timeline) )

  end
  span:finish()
  ngx.exit(ngx.HTTP_OK)
end

return _M