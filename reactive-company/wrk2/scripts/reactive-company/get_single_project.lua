local socket = require("socket")
math.randomseed(socket.gettime()*1000)
math.random(); math.random(); math.random()

local projects_no = 1

local projects_ids = {
    "62b6d1d8e1d71f0001a83727",
    "62b6d1d8e1d71f0001a83728",
    "62b6d1d8e1d71f0001a83729",
    "62b6d1d8e1d71f0001a83721",
  }


local function get_project()
    local method = "GET"
    local headers = {}
    headers["Content-Type"] = "application/json"

    local index = math.random(#projects_ids)
    local path = "http://localhost:8080/projects/" .. projects_ids[index]

    return wrk.format(method, path, headers, nil)
end

request = function()
    return get_project()
end
