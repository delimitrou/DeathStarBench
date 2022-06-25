local socket = require("socket")
math.randomseed(socket.gettime()*1000)
math.random(); math.random(); math.random()


local function get_all_projects()
    local method = "GET"
    local headers = {}

    local path = "http://localhost:8080/projects"

    return wrk.format(method, path, headers, nil)
end

request = function()
    return get_all_projects()
end
