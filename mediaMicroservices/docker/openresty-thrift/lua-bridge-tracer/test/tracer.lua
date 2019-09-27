local bridge_tracer = require 'opentracing_bridge_tracer'
local json = require 'JSON'

function new_mocktracer(json_file)
  local mocktracer_path = os.getenv("MOCKTRACER")
  if mocktracer_path == nil then
    error("MOCKTRACER environmental variable must be set")
  end
  return bridge_tracer:new(
            mocktracer_path, 
            '{ "output_file":"' .. json_file .. '" }')
end

function read_json(file)
    local f = assert(io.open(file, "rb"))
    local content = f:read("*all")
    f:close()
    return json:decode(content)
end

describe("in bridge_tracer", function()
  describe("tracer construction", function()
    it("fails when passed an invalid library", function()
      local errorfn = function()
        local tracer = bridge_tracer:new("invalid_library", "invalid_config")
      end
      -- Triggers ASAN false positive
      -- assert.has_error(errorfn)
    end)

    it("fails when passed an invalid tracer config", function()
      local mocktracer_path = os.getenv("MOCKTRACER")
      if mocktracer_path == nil then
        error("MOCKTRACER environmental variable must be set")
      end
      local errorfn = function()
        local tracer = bridge_tracer:new(mocktracer_path, "invalid_config")
      end
      -- Triggers ASAN false positive
      -- assert.has_error(errorfn)
    end)

    it("supports construction from a valid library and config", function()
      local json_file = os.tmpname()
      local tracer = new_mocktracer(json_file)
      assert.are_not_equals(tracer, nil)
    end)

    it("supports construction from the C++ global tracer", function()
      local tracer = bridge_tracer:new_from_global()
      assert.are_not_equals(tracer, nil)
    end)
  end)

  describe("the start_span method", function()
    it("creates spans", function()
      local json_file = os.tmpname()
      local tracer = new_mocktracer(json_file)
      local span = tracer:start_span("abc")
      span:finish()
      tracer:close()
			local json = read_json(json_file)
			assert.are.equal(#json, 1)
			assert.are.equal(json[1]["operation_name"], "abc")
    end)

    it("supports customizing the time points", function()
      local json_file = os.tmpname()
      local tracer = new_mocktracer(json_file)
      local span = tracer:start_span("abc", {["start_time"] = 1531434895308545})
      span:finish(1531434896813719)
      tracer:close()
			local json = read_json(json_file)
			assert.are.equal(#json, 1)
      assert.is_true(json[1]["duration"] > 1.0e6)
    end)

    it("errors when passed an incorrect operation name", function()
      local json_file = os.tmpname()
      local tracer = new_mocktracer(json_file)
      local errorfn = function()
        local span = tracer:start_span({})
      end
      assert.has_error(errorfn)
    end)

    it("supports referencing other spans", function()
      local json_file = os.tmpname()
      local tracer = new_mocktracer(json_file)
      local span_a = tracer:start_span("A")

      local span_b = tracer:start_span("B", {["references"] = {{"child_of", span_a:context()}}})
      span_b:finish()

      local span_c = tracer:start_span("C", {["references"] = {{"follows_from", span_a:context()}}})
      span_c:finish()

      span_a:finish()
      tracer:close()

			local json = read_json(json_file)
			assert.are.equal(#json, 3)
      local references_b = json[1]["references"]
			assert.are.equal(#references_b, 1)

      local references_c = json[2]["references"]
			assert.are.equal(#references_c, 1)
    end)

    it("ignore nil references", function()
      local json_file = os.tmpname()
      local tracer = new_mocktracer(json_file)
      local span = tracer:start_span("abc", {["references"] = {{"child_of", nil}}})
      span:finish()
      tracer:close()
			local json = read_json(json_file)
			assert.are.equal(#json, 1)
      local references = json[1]["references"]
			assert.are.equal(#references, 0)
    end);
  end)

  describe("a tracer", function()
    it("returns nil when extracting from an empty table", function()
      local json_file = os.tmpname()
      local tracer = new_mocktracer(json_file)
      -- text map
      local context1 = tracer:text_map_extract({})
      assert.are.equal(context1, nil)

      -- http headers
      local context2 = tracer:http_headers_extract({})
      assert.are.equal(context2, nil)

      -- binary
      local context3 = tracer:binary_extract("")
      assert.are.equal(context3, nil)
    end)
  end)

  describe("a span", function()
    it("supports changing the operation name", function()
      local json_file = os.tmpname()
      local tracer = new_mocktracer(json_file)
      local span = tracer:start_span("abc")
      span:set_operation_name("xyz")
      span:finish()
      tracer:close()
			local json = read_json(json_file)
			assert.are.equal(#json, 1)
      assert.are.equal(json[1]["operation_name"], "xyz")
    end)

    it("supports context propagation", function()
      local json_file = os.tmpname()
      local tracer = new_mocktracer(json_file)
      local span = tracer:start_span("abc")
      span:finish()

      -- text map
      local carrier1 = {}
      tracer:text_map_inject(span:context(), carrier1)
      local context1 = tracer:text_map_extract(carrier1)
      assert.are_not_equals(context1, nil)

      -- http headers
      local carrier2 = {}
      tracer:http_headers_inject(span:context(), carrier2)
      local context2 = tracer:http_headers_extract(carrier2)
      assert.are_not_equals(context2, nil)

      -- binary
      local carrier3 = tracer:binary_inject(span:context())
      local context3 = tracer:binary_extract(carrier3)
      assert.are_not_equals(context3, nil)

      -- ignores non-string key-value pairs
      local tbl = {}
      local carrier4 = {["k1"] = "v1", [tbl] = "abc", ["abc"] = tbl}
      local context4 = tracer:text_map_extract(carrier4)
      assert.are.equal(context4, nil)
    end)

    it("supports obtaining a reference to the tracer that created it", function()
      local json_file = os.tmpname()
      local tracer = new_mocktracer(json_file)
      local span = tracer:start_span("abc")
      local tracer2 = span:tracer()
      assert.are_not_equals(tracer2, nil)
    end)

    it("is correctly garbage collected", function()
      local json_file = os.tmpname()
      local tracer = new_mocktracer(json_file)
      local span = tracer:start_span("abc")
      local context = span:context()
      tracer = nil
      span = nil
      context = nil

      -- free functions should be called for the tracer, span, and context
      -- 
      -- when run with address sanitizer, leaks should be detected if they
      -- aren't freed
      collectgarbage()
    end)

    it("supports attaching tags", function()
      local json_file = os.tmpname()
      local tracer = new_mocktracer(json_file)
      local span = tracer:start_span("abc")
      span:set_tag("s", "abc")
      span:set_tag("i", 123)
      span:finish()
      tracer:close()
			local json = read_json(json_file)
			assert.are.equal(#json, 1)
			assert.are.equal(json[1]["tags"]["s"], "abc")
			assert.are.equal(json[1]["tags"]["i"], 123)
    end)

    it("supports logging", function()
      local json_file = os.tmpname()
      local tracer = new_mocktracer(json_file)
      local span = tracer:start_span("abc")
      span:log_kv({["x"] = 123})
      span:finish()
      tracer:close()
			local json = read_json(json_file)
			assert.are.equal(#json, 1)
      local logs = json[1]["logs"]
      assert.are.equal(#logs, 1)
      local records = logs[1]["fields"]
      assert.are.equal(#records, 1)
      assert.are.equal(records[1]["key"], "x")
      assert.are.equal(records[1]["value"], 123)
    end)

    it("supports attaching and querying baggage", function()
      local json_file = os.tmpname()
      local tracer = new_mocktracer(json_file)
      local span = tracer:start_span("abc")
      span:set_baggage_item("abc", "123")
      assert.are_equal(span:get_baggage_item("abc"), "123")
    end)

  end)
end)
