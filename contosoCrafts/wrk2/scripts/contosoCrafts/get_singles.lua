local socket = require("socket")
math.randomseed(socket.gettime()*1000)
math.random(); math.random(); math.random()

local products_no = 15

local function get_product()
    local method = "GET"
    local headers = {}
    headers["Content-Type"] = "application/json"

    local id = math.random(0, products_no)
    local path = "http://localhost:9090/Products/Index=" .. id

    return wrk.format(method, path, headers, nil)
end

request = function()
    return get_product()
end
