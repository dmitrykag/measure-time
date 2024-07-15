-- Initialize the pseudo random number generator
-- Resource: http://lua-users.org/wiki/MathLibraryTutorial
math.randomseed(os.time())
math.random(); math.random(); math.random()

-- Shuffle array
-- Returns a randomly shuffled array
function shuffle(paths)
  local j, k
  local n = #paths

  for i = 1, n do
    j, k = math.random(n), math.random(n)
    paths[j], paths[k] = paths[k], paths[j]
  end

  return paths
end

-- Load URL paths from the file
function load_request_objects_from_file(file)
  local lines = {}

  -- Check if the file exists
  -- Resource: http://stackoverflow.com/a/4991602/325852
  local f=io.open(file,"r")
  if f~=nil then
    for line in io.lines(file) do
      lines[#lines + 1] = line
    end
   
    io.close(f)
  else
    -- Return the empty array
    return lines
  end

  return shuffle(lines)
end

queries = load_request_objects_from_file("queries.txt")

-- Check if at least one path was found in the file
if #queries <= 0 then
  print("multiplerequests: No requests found.")
  os.exit()
end

print("multiplerequests: Found " .. #queries .. " requests")

-- Initialize the requests array iterator
counter = 1

wrk.method = "POST"
wrk.headers["Content-Type"] = "application/json"

request = function()
  -- Get the next requests array element
  local query_object = queries[counter]

  -- Increment the counter
  counter = counter + 1

  -- If the counter is longer than the requests array length then reset it
  if counter > #queries then
    counter = 1
  end

  -- Return the request object with the current URL path
  return wrk.format(nil, nil, nil, "{\"query\": \"" .. query_object .. "\"}")
end
