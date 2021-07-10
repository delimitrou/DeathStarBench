counter = 0

-- Initialize the pseudo random number generator - http://lua-users.org/wiki/MathLibraryTutorial
math.randomseed(os.time())
math.random(); math.random(); math.random()

function file_exists(file)
  local f = io.open(file, "rb")
  if f then f:close() end
  return f ~= nil
end

function shuffle(paths)
  local j, k
  local n = #paths
  for i = 1, n do
    j, k = math.random(n), math.random(n)
    paths[j], paths[k] = paths[k], paths[j]
  end
  return paths
end

function non_empty_lines_from(file)
  if not file_exists(file) then return {} end
  lines = {}
  for line in io.lines(file) do
    if not (line == '') then
      lines[#lines + 1] = line
    end
  end
  return shuffle(lines)
end

paths = non_empty_lines_from("paths.txt")

if #paths <= 0 then
  print("multiplepaths: No paths found. You have to create a file paths.txt with one path per line")
  os.exit()
end

print("multiplepaths: Found " .. #paths .. " paths")

init = function(args)
    depth = tonumber(args[1]) or 1
end

request = function()
    path = paths[counter]
    counter = counter + 1
    if counter > #paths then
      counter = 0
    end
--    return wrk.format(nil, path)
    
--    depth = tonumber(args[1]) or 1
    local r = {}
    for i=1, depth do
        r[i] = wrk.format(nil, path)
    end
    req = table.concat(r)
    return req
end
