local _M = {}
local k8s_suffix = os.getenv("fqdn_suffix")
if (k8s_suffix == nil) then
  k8s_suffix = ""
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

function _M.UploadMedia()
  local upload = require "resty.upload"
  local mongo = require "resty-mongol"
  local cjson = require "cjson"
  local ngx = ngx

  local chunk_size = 8196
  local form, err = upload:new(chunk_size)
  if not form then
    ngx.log(ngx.ERR, "failed to new upload: ", err)
    ngx.exit(500)
  end

  form:set_timeout(1000)
  local media_id = tonumber(string.sub(ngx.var.request_id, 0, 15), 16)
  media_id = string.format("%.f", media_id)
  local media_file = ""
  local media_type

  while true do
    local typ, res, err = form:read()
    if not typ then
      ngx.say("failed to read: ", err)
      return
    end

    if typ == "header" then
      for i, ele in ipairs(res) do
        local filename = string.match(ele, 'filename="(.*)"')
        if filename and filename ~= '' then
          local filename_list = _StringSplit(filename, '.')
          media_type = filename_list[#filename_list]
        end
      end
    elseif typ == "body" then
      media_file = media_file .. res
    elseif typ == "part_end" then

    elseif typ == "eof" then
      break
    end
  end

  local conn = mongo()
  conn:set_timeout(1000)
  local ok, err = conn:connect("media-mongodb" .. k8s_suffix, 27017)
  if not ok then
    ngx.log(ngx.ERR, "mongodb connect failed: "..err)
  end
  local db = conn:new_db_handle("media")
  local col = db:get_col("media")

  local media = {
    filename = media_id .. '.' ..  media_type,
    file = media_file
  }
  col:insert({media})
  conn:set_keepalive(60000, 100)
  ngx.header.content_type = "application/json; charset=utf-8"
  ngx.say(cjson.encode({media_id = media_id, media_type = media_type}))

end

return _M