local socket = require("socket")
math.randomseed(socket.gettime()*1000)

function get_random_string(string_length)
    local result_string = ""
    for i = 1, string_length do
        result_string = result_string .. string.char(math.random(97, 122))
    end
    return result_string
end

function add_blogpost()
    local title_length = math.random(5,10)
    local title = get_random_string(title_length)
    local random_published_value = math.random(1,2) == 1
    local published = tostring(random_published_value)


    local method = "POST"
    local path = "http://localhost:8080/blogposts?title=" .. title .. "&published=" .. published
    local headers = {}
    return wrk.format(method, path, headers, nil)
end


local function get_all()
    local method = "GET"
    local headers = {}
    headers["Content-Type"] = "application/json"

    local path = "http://localhost:8080/"

    return wrk.format(method, path, headers, nil)
end


local function get_all_blogposts_stream()
    local method = "GET"
    local headers = {}
    headers["Content-Type"] = "application/json"

    --choose endpoint for stream 
    local path = "http://localhost:8080/stream"
    -- local path = "http://localhost:8080/stream/blog"
    return wrk.format(method, path, headers, nil)
end


iteration = -1

request = function()
    -- SET to 1 to get only a prints higher to set some posts beetween
    local iteration_size_update_numer = 4
    
    iteration = iteration + 1
    if iteration >  iteration_size_update_numer then
        iteration = 0
    end
    if iteration == 0 then
      return get_all()
    elseif iteration == 1  then
        return get_all_blogposts_stream()
    else

        return add_blogpost()
    end
end

last_normal = 0
last_stream = 0

-- CHECKIN RESPONSE WITH STATUS , len of body and change of body
response = function(status, headers, body)
    if iteration == 0 then
        print(" Blog normal more :" .. #body - last_normal )
        last_normal = #body
        return print(" Blog normal SIZE :" .. #body )
      elseif iteration == 1  then
        print("status" .. status .. "Blog stream more :" .. #body - last_stream )
        last_stream= #body
        return print(" Blog stream SIZE :" .. #body)
    end
end

-- CHECKIN RESPONSE WITH STATUS , time and len of body
-- response = function(status, headers, body)
--     if iteration == 0 then
--         return print("status: " .. status .. " Blog normal SIZE " .. #body .. " time ".. os.time())
--       elseif iteration == 1  then
--         return print("status: " .. status ..  " Blog stream SIZE" .. #body .. " time ".. os.time())
--     end
-- end
