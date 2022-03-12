// 你可以使用该函数主动崩溃脚本
engine.crash("在这里脚本崩溃了!")

// 当重复使用 script 指令加载同一个脚本时，前一个会被停止
// script example.js // 第一次加载
// 修改 example.js
// script example.js // 第二次加载时，第一次加载的脚本会被终止

