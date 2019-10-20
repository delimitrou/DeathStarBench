require "socket"
time = socket.gettime()*1000
math.randomseed(time)
math.random(); math.random(); math.random()

local charset = {'q', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o', 'p', 'a', 's',
  'd', 'f', 'g', 'h', 'j', 'k', 'l', 'z', 'x', 'c', 'v', 'b', 'n', 'm', 'Q',
  'W', 'E', 'R', 'T', 'Y', 'U', 'I', 'O', 'P', 'A', 'S', 'D', 'F', 'G', 'H',
  'J', 'K', 'L', 'Z', 'X', 'C', 'V', 'B', 'N', 'M', '1', '2', '3', '4', '5',
  '6', '7', '8', '9', '0'}

local movie_titles = {
  "Avengers: Endgame",
  "Kamen Rider Heisei Generations FOREVER",
  "Captain Marvel",
  "Pokémon Detective Pikachu",
  "Hellboy",
  "After",
  "Avengers: Infinity War"
}

function string.random(length)
  if length > 0 then
    return string.random(length - 1) .. charset[math.random(1, #charset)]
  else
    return ""
  end
end

request = function()
  local movie_index = math.random(7)
  local user_index = math.random(1000)
  local username = "username_" .. tostring(user_index)
  local password = "password_" .. tostring(user_index)
  local title = movie_titles[movie_index]
  local rating = math.random(0, 10)
  local text = string.random(256)

  local path = "http://10.68.43.165:8080/wrk2-api/review/compose"
  local method = "POST"
  local headers = {}
  local body = "username=" .. username .. "&password=" .. password .. "&title=" ..
                  title .. "&rating=" .. rating .. "&text=" .. text
  headers["Content-Type"] = "application/x-www-form-urlencoded"

  return wrk.format(method, path, headers, body)
end