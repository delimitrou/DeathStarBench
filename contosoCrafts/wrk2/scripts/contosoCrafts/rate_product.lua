local socket = require("socket")
math.randomseed(socket.gettime()*1000)
math.random(); math.random(); math.random()

local product_ids = {
  "jenlooper-cactus",
  "jenlooper-light",
  "jenlooper-lightshow",
  "jenlooper-survival",
  "sailorhg-bubblesortpic",
  "sailorhg-corsage",
  "sailorhg-kit",
  "sailorhg-led",
  "selinazawacki-soi-shirt",
  "selinazawacki-soi-pins",
  "vogueandcode-hipster-dev-bro",
  "vogueandcode-pretty-girls-code-tee",
  "vogueandcode-ruby-sis-2",
  "selinazawacki-moon",
  "selinazawacki-shirt",
}

function urlEncode(s)
     s = string.gsub(s, "([^%w%.%- ])", function(c) return string.format("%%%02X", string.byte(c)) end)
    return string.gsub(s, " ", "+")
end

function rate_product()
    local random_id = math.random(#product_ids)
    local product_id = urlEncode(product_ids[random_id])
    local rating = math.random(1, 5)

    local method = "PATCH"    
    local headers = {}
    local body = '{"productId": "' .. product_id .. '", "rating": ' .. rating .. '}'
    local path = "http://localhost:9090/Products/"
    headers["Content-Type"] = "application/json"
    
    return wrk.format(method, path, headers, body)
end

request = function()
    return rate_product()
end
