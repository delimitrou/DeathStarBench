--
-- Licensed to the Apache Software Foundation (ASF) under one
-- or more contributor license agreements. See the NOTICE file
-- distributed with this work for additional information
-- regarding copyright ownership. The ASF licenses this file
-- to you under the Apache License, Version 2.0 (the
-- "License"); you may not use this file except in compliance
-- with the License. You may obtain a copy of the License at
--
--   http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing,
-- software distributed under the License is distributed on an
-- "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
-- KIND, either express or implied. See the License for the
-- specific language governing permissions and limitations
-- under the License.
--

local TSocket = require 'TSocket'
local Thrift = require 'Thrift'
local TTransport = require 'TTransport'
local TTransportException = TTransport.TTransportException
local ttype = Thrift.ttype
local terror = Thrift.terror

-- TSocketSSL
local TSocketSSL = TSocket:new{
  __type = 'TSocketSSL',
  host = 'localhost',
  port = 9090
}

function TSocketSSL:open()
  ngx.log(ngx.INFO, "ssl open called")
  if not self.handle then
    -- Use NGINX socket instead of the built-in lua socket
    self.handle = ngx.socket.tcp()
    self.handle:settimeout(self.timeout)
  end
  ngx.log(ngx.INFO, "ssl start to connect")
  local ok, err = self.handle:connect(self.host, self.port)
  if not ok then
    terror(TTransportException:new{
      message = 'Could not connect to ' .. self.host .. ':' .. self.port
        .. ' (' .. err .. ')'
    })
  end
  ngx.log(ngx.INFO, "ssl handshake start")
  local session, err = self.handle:sslhandshake(nil, nil, true)
  if not session then
    terror(TTransportException:new{
      message = 'failed to do hand shake with ' .. self.host .. ':' .. self.port
        .. ' (' .. err .. ')'
    })
  end
  ngx.log(ngx.INFO, "ssl handshake end")
end

return TSocketSSL
