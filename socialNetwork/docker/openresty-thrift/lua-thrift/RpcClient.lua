--
----Author: xiajun
----Date: 20151020
----
local TSocket = require "TSocket"
local TFramedTransport = require "TFramedTransport"
local TBinaryProtocol = require "TBinaryProtocol"
local Object = require "Object"

local RpcClient = Object:new({
	__type = 'RpcClient',
})

--初始化RPC连接
function RpcClient:init(ip,port,timeout)
	local socket = TSocket:new{
		host = ip,
		port = port,
	}
	socket:setTimeout(timeout)
	local transport = TFramedTransport:new{
		trans = socket
	}
	local protocol = TBinaryProtocol:new{
		trans = transport
	}
	transport:open()
	return protocol;
end
--创建RPC客户端
function RpcClient:createClient(thriftClient)end

return RpcClient
