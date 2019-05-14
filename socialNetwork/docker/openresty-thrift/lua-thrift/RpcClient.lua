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
	timeout = 1001,
	readTimeout = 500
})

--初始化RPC连接
function RpcClient:init(ip,port)
	local socket = TSocket:new{
		host = ip,
		port = port,
		readTimeout = self.readTimeout
	 }
	socket:setTimeout(self.timeout)
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
