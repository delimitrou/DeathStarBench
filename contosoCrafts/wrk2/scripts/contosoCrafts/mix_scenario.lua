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

local products_no = 15

local function get_product()
    local method = "GET"
    local headers = {}
    headers["Content-Type"] = "application/json"

    local id = math.random(0, products_no)
    local path = "http://localhost:9090/Products/Index=" .. id

    return wrk.format(method, path, headers, nil)
end


local function get_all_products()
    local method = "GET"
    local headers = {}
    headers["Content-Type"] = "application/json"

    local path = "http://localhost:9090/Products"

    return wrk.format(method, path, headers, nil)
end

request = function()
    cur_time = math.floor(socket.gettime())
    local rate_ratio      = 0.5
    local get_single_ratio   = 0.3
    local get_all_ratio   = 0.2

  
    local coin = math.random()
    if coin < rate_ratio then
      return rate_product()
    elseif coin < rate_ratio + get_single_ratio then
        return get_product()
    else
        return get_all_products()
    end
end