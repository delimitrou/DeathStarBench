local socket = require("socket")
math.randomseed(socket.gettime()*1000)
math.random(); math.random(); math.random()


local function get_all_products()
    local method = "GET"
    local headers = {}
    headers["Content-Type"] = "application/json"

    local path = "http://localhost:9090/Products"

    return wrk.format(method, path, headers, nil)
end

request = function()
    return get_all_products()
end


done = function(summary, latency, requests)
    local log = "%f,%f,%f,%d,%d,%d"
    print(log:format(latency.min, latency.mean, latency.max, summary.errors.status, summary.errors.timeout, summary.errors.connect))
end
