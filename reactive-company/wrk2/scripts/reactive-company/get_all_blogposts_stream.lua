local socket = require("socket")
math.randomseed(socket.gettime()*1000)


local function get_all_blogposts_stream()
    local method = "GET"
    local headers = {}
    headers["Content-Type"] = "application/json"

    --choose endpoint for stream
    local path = "http://localhost:8080/stream"
    -- local path = "http://localhost:8080/stream/blog"
    return wrk.format(method, path, headers, nil)
end


request = function()
    return get_all_blogposts_stream()
end


done = function(summary, latency, requests)
    local log = "%f,%f,%f,%d,%d,%d"
    print(log:format(latency.min, latency.mean, latency.max, summary.errors.status, summary.errors.timeout, summary.errors.connect))
end