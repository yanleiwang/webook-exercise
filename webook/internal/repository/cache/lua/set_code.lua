
local key = KEYS[1]
-- 验证次数，我们一个验证码，最多可以验证 3 次，这个记录还可以验证几次
local keyCnt = key..":cnt"
-- 验证码
local code = ARGV[1]

local ttl = tonumber(redis.call("ttl", key))
if ttl == -1 then
    --    key 存在，但是没有过期时间
    -- 系统错误，你的同事手贱，手动设置了这个 key，但是没给过期时间
    return -2

elseif ttl == -2 or ttl < 540  then
    -- key 不存在 或者 过期时间  <  540 = 600 -60 = 9分钟
    -- 可以设置验证码
    redis.call("set", key, code, "ex", expiration)
    redis.call("set", keyCnt, maxCnt, "ex", expiration)
    return 0
else
    -- 发送太频繁
    return -1
end
