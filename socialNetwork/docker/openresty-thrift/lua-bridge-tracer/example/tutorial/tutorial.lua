#!/usr/bin/env lua
--[[
Provides an example of basic usage of the opentracing_bridge_tracer.

Usage: lua tutorial.lua <path-to-tracer-plugin> <path-to-tracer-config>
]]

local bridge_tracer = require 'opentracing_bridge_tracer'

assert(#arg == 2)

-- initialize a tracer
local tracer_library = arg[1]
local config_file = arg[2]

local f = assert(io.open(config_file, "rb"))
local config = f:read("*all")
f:close()

local tracer = bridge_tracer:new(tracer_library, config)

local parent_span = tracer:start_span("parent")

-- create a child span
local child_span = tracer:start_span(
                    "ChildA",
                    {["references"] = {{"child_of", parent_span:context()}}})

child_span:set_tag("simple tag", 123)
child_span:log_kv({["event"] = "simple log", ["abc"] = 123})
child_span:finish()

-- propagate the span context
local carrier = {}
tracer:text_map_inject(parent_span:context(), carrier)
local span_context = tracer:text_map_extract(carrier)
assert(span_context ~= nil)

local propagation_span = tracer:start_span(
                      "PropagationSpan",
                      {["references"] = {{"follows_from", span_context}}})
propagation_span:finish()

-- close the tracer to ensure that spans are flushed
parent_span:finish()
tracer:close()
