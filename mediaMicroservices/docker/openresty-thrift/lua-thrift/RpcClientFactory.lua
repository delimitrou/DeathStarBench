--
----Author: xiajun
----Date: 20151020
----
local RpcClient = require "RpcClient"
--local TestServiceClient = require "resty.thrift.thrift-idl.lua_test_TestService"
local RpcClientFactory = RpcClient:new({
	__type = 'Client'
})
function RpcClientFactory:createClient(thriftClient, ip, port)
    local protocol = self:init(ip, port)
    local client = thriftClient:new{
        iprot = protocol,
        oprot = protocol
    }
    return client
end
return RpcClientFactory
