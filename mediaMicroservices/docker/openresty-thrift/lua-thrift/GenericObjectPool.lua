--
--Author: xiajun
--Date: 20151120
--
local Object = require 'Object'
local RpcClientFactory = require 'RpcClientFactory'
local ngx = ngx
local GenericObjectPool = Object:new({
    __type = 'GenericObjectPool',
    maxTotal = 100,
    maxIdleTime = 60000
    })
function GenericObjectPool:init(conf)
end
--
--从连接池获取rpc客户端
--ngx nginx容器变量
--
function GenericObjectPool:connection(thriftClient,ip,port)
    local client = RpcClientFactory:createClient(thriftClient,ip,port)
    return client
end
--
--回收连接资源到连接池
--client rpc客户端对象
--
function GenericObjectPool:returnConnection(client)
    if(client ~= nil)then
        if (client.iprot.trans.trans:isOpen())then
            client.iprot.trans.trans:setKeepAlive(self.maxIdleTime, self.maxTotal)
        else
            ngx.log(ngx.ERR,"return rpc client fail ,socket close.")
        end
    end
end
--
--设置连接池的大小
--Maxtotal 连接池大小
--
function GenericObjectPool:setMaxTotal(maxTotal)
  self.maxTotal = maxTotal
end
function GenericObjectPool:clear()

end
function GenericObjectPool:remove()

end
return GenericObjectPool
