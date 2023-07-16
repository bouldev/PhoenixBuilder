local omega = require("omega")
local print = omega.system.print
local save_data = omega.storage.save     --   保存数据
local read_data = omega.storage.read     -- 读取数据
local remove_data = omega.storage.remove --移除数据文件

-- 准备一些数据用来测试
local data = {
    name = "omega",
    version = "0.0.1",
    author = "omega",
    description = "omega lua api",
    keywords = { "omega", "lua", "api" },
    license = "MIT",
    homepage = "",
    some_number = 123,
    subset = {
        type = "lua",
        disigner = "omega"
    }
}

-- 设定一个文件，用来保存和读取数据
local file = "test_data.json"

-- 保存数据
save_data(file, data)

-- 读取数据
local data_read = read_data(file)
print(("data read: %s"):format(data_read["name"]))
print(("data read: %s"):format(data_read["some_number"]))
print(("data read: %s"):format(data_read["keywords"]))
for i,v in ipairs(data_read["keywords"]) do
    print(("\t%s:%s"):format(i,v))
end
print(("data read: %s"):format(data_read["subset"]))
for i,v in pairs(data_read["subset"]) do
    print(("\t%s:%s"):format(i,v))
end

-- 移除数据文件
remove_data(file)