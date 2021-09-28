# lua-bridge-tracer

Provides an implementation of the [Lua OpenTracing API](https://github.com/opentracing/opentracing-lua)
on top of the [C++ OpenTracing API](https://github.com/opentracing/opentracing-cpp).

Dependencies
------------
- The [C++ OpenTracing Library](https://github.com/opentracing/opentracing-cpp)
- A C++ OpenTracing Tracer. It currently works with
[Jaeger](https://github.com/jaegertracing/cpp-client),
[Zipkin](https://github.com/rnburn/zipkin-cpp-opentracing), or
[LightStep](https://github.com/lightstep/lightstep-tracer-cpp).

Installation
------------
```bash
mkdir .build
cd .build
cmake ..
make
sudo make install
```

Usage
-----
```lua
bridge_tracer = require 'opentracing_bridge_tracer'
library = --[[ path to OpenTracing plugin ]]
config = --[[ vendor specific JSON configuration for the tracer ]]
tracer = bridge_tracer:new(library, config)

-- `tracer` conforms to the Lua OpenTracing API. See
-- https://github.com/opentracing/opentracing-lua for API documentation.
```

See also [example/tutorial](example/tutorial).
