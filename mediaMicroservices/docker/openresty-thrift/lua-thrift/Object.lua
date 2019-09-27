--
----Author: xiajun
----Date: 20151120
----
local function ttype(obj)
    if type(obj) == 'table' and obj.__type and type(obj.__type) == 'string' then
        return obj.__type
    end
    return type(obj)
end
local function __obj_index(self, key)
    local obj = rawget(self, key)
    if obj ~=nil then
        return obj
    end
    local p = rawget(self,'__parent')
    if p then
        return __obj_index(p,key)
    end
    return nil
end

local Object = {
    __type = 'Object',
    __mt = {
        __index = __obj_index
    }
}
function Object:new(init_obj)
    local obj = {}
    if ttype(obj) == 'table' then
        obj = init_obj
    end
    obj.__parent = self
    setmetatable(obj,Object.__mt)
    return obj
end
function Object:ttype()
    if self and self.__type then
        return self.__type
    end
    return type(self)
end
return Object
