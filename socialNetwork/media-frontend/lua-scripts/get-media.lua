local _M = {}
local k8s_suffix = os.getenv("fqdn_suffix")
if (k8s_suffix == nil) then
  k8s_suffix = ""
end

local function _StrIsEmpty(s)
  return s == nil or s == ''
end

local function _StringSplit(input_str, sep)
  if sep == nil then
    sep = "%s"
  end
  local t = {}
  for str in string.gmatch(input_str, "([^"..sep.."]+)") do
    table.insert(t, str)
  end
  return t
end

function _M.GetMedia()
  local mongo = require "resty-mongol"
  local ngx = ngx

  local chunk_size = 255 * 1024

  ngx.req.read_body()
  local args = ngx.req.get_uri_args()
  if (_StrIsEmpty(args.filename)) then
    ngx.status = ngx.HTTP_BAD_REQUEST
    ngx.say("Incomplete arguments")
    ngx.log(ngx.ERR, "Incomplete arguments")
    ngx.exit(ngx.HTTP_BAD_REQUEST)
  end


  local conn = mongo()
  conn:set_timeout(1000)
  local ok, err = conn:connect("media-mongodb" .. k8s_suffix, 27017)
  if not ok then
    ngx.log(ngx.ERR, "mongodb connect failed: "..err)
  end
  local db = conn:new_db_handle("media")
  local col = db:get_col("media")

  local media = col:find_one({filename=args.filename})
  if not media then
    ngx.log(ngx.ERR, "mongodb failed to find: ".. args.filename)
    return
  end

  local media_file = media.file

  local filename_list = _StringSplit(args.filename, '.')
  local media_type = filename_list[#filename_list]

  ngx.header.content_type = "image/" .. media_type
  ngx.say(media_file)

end

return _M