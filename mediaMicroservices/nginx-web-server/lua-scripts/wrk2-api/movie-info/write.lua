local _M = {}
local k8s_suffix = os.getenv("fqdn_suffix")
if (k8s_suffix == nil) then
  k8s_suffix = ""
end

function _M.WriteMovieInfo()
  local bridge_tracer = require "opentracing_bridge_tracer"
  local GenericObjectPool = require "GenericObjectPool"
  local MovieInfoServiceClient = require 'media_service_MovieInfoService'
  local ttypes = require("media_service_ttypes")
  local Cast = ttypes.Cast
  local ngx = ngx
  local cjson = require("cjson")

  local req_id = tonumber(string.sub(ngx.var.request_id, 0, 15), 16)
  local tracer = bridge_tracer.new_from_global()
  local parent_span_context = tracer:binary_extract(ngx.var.opentracing_binary_context)
  local span = tracer:start_span("WriteMovieInfo", {["references"] = {{"child_of", parent_span_context}}})
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

  local movie_info = cjson.decode(data)
  if (movie_info["movie_id"] == nil or movie_info["title"] == nil or
      movie_info["casts"] == nil or movie_info["plot_id"] == nil or
      movie_info["thumbnail_ids"] == nil or movie_info["photo_ids"] == nil or
      movie_info["video_ids"] == nil or movie_info["avg_rating"] == nil or
      movie_info["num_rating"] == nil) then
    ngx.status = ngx.HTTP_BAD_REQUEST
    ngx.say("Incomplete arguments")
    ngx.log(ngx.ERR, "Incomplete arguments")
    ngx.exit(ngx.HTTP_BAD_REQUEST)
  end

  local casts = {}
  for _,cast in ipairs(movie_info["casts"]) do
    local new_cast = Cast:new{}
    new_cast["charactor"]=cast["charactor"]
    new_cast["cast_id"]=cast["cast_id"]
    new_cast["cast_info_id"]=cast["cast_info_id"]
    table.insert(casts, new_cast)
  end


  local client = GenericObjectPool:connection(MovieInfoServiceClient, "movie-info-service" .. k8s_suffix , 9090)
  client:WriteMovieInfo(req_id, movie_info["movie_id"], movie_info["title"],
      casts, movie_info["plot_id"], movie_info["thumbnail_ids"],
      movie_info["photo_ids"], movie_info["video_ids"], tostring(movie_info["avg_rating"]),
      movie_info["num_rating"], carrier)
  ngx.say(movie_info["avg_rating"])
  GenericObjectPool:returnConnection(client)

end

return _M