local omega = require("omega")
local print = omega.system.print

-- 先用json的形式看一眼所有配置
print(omega.config:json_str())

-- 读取配置中我们需要的项目
local config = omega.config:user_data() -- 将配置转为userdata形式
local version = config.Version        -- 可以直接读取和写入数据
print(("version: %s"):format(version))
config.Version = "0.0.1"              -- 修改配置

-- 然而，user data 有个小麻烦，就是它无法像正常的lua table 一样遍历
-- 这个时候，就需要将其通过 ud2lua 转为正常的 table
-- ud2lua 可以应用于任何user data 上
local users = config.Users
for user, role in pairs(ud2lua(users)) do
    print(("%s : %s"):format(user, role))
end

omega.config:upgrade(config) -- 升级配置

-- 或者也可以这样
local config = ud2lua(omega.config:user_data())
print(("%s"):format(config))
print(("version %s"):format(config["Version"]))
print(("user %s"):format(config["Users"]))
for user, role in pairs(config["Users"]) do
    print(("%s : %s"):format(user, role))
end

config["Version"] = "0.0.3"
omega.config:upgrade(config)

-- 选择那种风格，这种事情还是看个人喜好吧
