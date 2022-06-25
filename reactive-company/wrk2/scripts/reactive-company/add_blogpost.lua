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

request = function()
    return add_blogpost()
end
