# BDump文件格式

> 感谢 [EillesWan](https://github.com/EillesWan) 提供的翻译，本文档内容主要来自其翻译内容，存在部分改动。
>
> 所有注有`译注`的内容均为以上贡献者所注。

BDump v3 是个用于存储*Minecraft*建筑结构的文件格式。其内容由指示建造过程的命令组成。

按照一定的顺序来写下每一个方块的ID的文件格式会因为包含空气方块而徒增文件大小，因此我们设计了一种新的文件格式，引入了「画笔」，并让一系列的指令控制其进行移动或放置方块。

*\[注：画笔绝非机器人的位置，而是一个引入的抽象的概念\]*

## 基本文件结构

BDump v3 文件的后缀名为`.bdx`，且文件头为`BD@`, 代表本bdump文件已使用 brotli（PhoenixBuilder 使用的压缩质量为`6`）进行压缩。请注意，文件头为`BDZ`的 BDump 文件同时存在，其使用 gzip 压缩，然而包含这种文件头的`.bdx`文件不被 PhoenixBuilder 支持，因为其弃用较早，目前难以再找到此类型的文件。我们将这种文件头定义为“压缩头”(compression header)，并且在此压缩头后面的内容将以压缩头所表明的方式进行压缩。

> 注: BDump v2 的文件后缀是 `.bdp`，且文件头为 `BDMPS\0\x02\0`。

在压缩头之后的，压缩后内容的起始字符为 `BDX\0`，且作者的游戏名紧跟其后，并以 `\0` 表示其玩家名的表示完毕。*\[译注：即若作者之游戏名为Eilles，则此文件压缩后的内容应以*`BDX\0Eilles\0`*开头\]* 此后之文本即含参指令了，它们惜惜相依，紧紧相连。每个指令的ID占有1字节(byte)的空间，其正是`unsigned char`所占的空间。

所有的操作都基于一个用以标识“画笔”所在位置的 `Vec3` 值。

*\[译注：原谅我才疏学浅，冒昧在这里注明一下：* `Vec3` *值指的是一个用以表示三维向量或坐标的值\]*

来来来，我们看看指令列表先。

数据类型定义如下：

* {整型}(int)：即全体整数集，可包含正整数、0、负整数
* {无符号整型}（亦称非负整型）(unsigned int)：即全体非负整数集，可包含正整数和0
* `char`(单字)（亦称字符）：一个1字节长的{整型}值
* `unsigned char`(无符号单字)（亦称无符号字符或非负字符）：一个1字节长的{无符号整型}值
* `short`(短整)：一个2字节长的{整型}值
* `unsigned short`(无符号短整（亦称非负短整）)：一个2字节长的{无符号整型}值
* `int32_t`：4字节长的{整型}数据
* `uint32_t`：4字节长的{无符号整型}数据
* `char *`：以`\0`(UTF-8编码)结尾的字符串
* `int`：即`int32_t`
* `unsigned int`：即`uint32_t`
* `bool`(布尔)：1字节长的布尔(亦称逻辑)数据，仅可为真(`true`, `1`)或假(`false`, `0`)

> 请注意：BDump文件的数字信息将会以<font style="color:red;">**大端字节序**</font>(big endian)又称<font style="color:red;">**大端序**</font>记录.
>
> 大小端字节序有何不同呢？
>
> *\[译注：你完全可以去查百度、必应上面搜索出来的解析，那玩意肯定让你半蒙半懂，但这玩意本身相对而言也并非十分绝对得重要，你看下面这个全蒙的也挺好。\]*
>
> 例如，一个`int32`的`1`在小端字节序的表示下，内存中是这样的`01 00 00 00`，而大端为`00 00 00 01`。

*\[译注：下面这表格中，我把调色板(palette)翻译为了方块池，纯是因为意译，但是，我也知道这样失去了很多原文的趣味，我也在思索一种更好的翻译……\]*

| ID                | 内部名                                     | 描述                                                         | 参数                                                         |
| ----------------- | ------------------------------------------ | ------------------------------------------------------------ | ------------------------------------------------------------ |
| 1                 | `addToBlockPalette`                        | 将所需方块名加入方块池中，且这神奇的方块池依照你调用这个命令(`addToBlockPalette`)的顺序为你可爱的方块名称分配ID。就是说，你第一次调用这个命令的时候，方块的id为`0`，第二次就是`1`了。哦，我的天哪，最多的方块id可以达到`65535`个之多！ | `char *blockName`                                            |
| 2                 | `addX`                                     | **(已弃用)** 将画笔的 `X` 坐标增加 `x`，顺带把画笔的 `Y` 和 `Z` 坐标重置为 `0`。 由于命令的实际功能与其名称之间存在差异，该方法在我们生成过程中不会再使用。虽然它已被我们弃用，但在读取 `bdx` 时仍然需要实现它的解析，因为包含此命令的 `bdx` 文件，也就是旧版本的文件一直都在。 | `unsigned short x`                                           |
| 3                 | `X++`                                      | **(已弃用)** 将画笔的 `X` 坐标增加 `1`，顺带把画笔的 `Y` 和 `Z` 坐标重置为 `0`。（同上理） | -                                                            |
| 4                 | `addY`                                     | **(已弃用)** 将画笔的 `Y` 坐标增加 `y`，顺带把画笔的 `Z` 坐标重置为 `0`。（同上理） | `unsigned short y`                                           |
| 5                 | `Y++`                                      | **(已弃用)** 将画笔的 `Y` 坐标增加 `1`，顺带把画笔的 `Z` 坐标重置为 `0`。（同上理） | -                                                            |
| 6                 | `addZ`                                     | 将画笔的 `Z` 坐标增加 `z`，哦我的天哪，这竟然并没有被弃用，那是因为它并不会把什么东西搞没；可是，理所应当的，这玩意也不再在当下版本的 PhonixBuilder 输出的文件中被使用了。 | `unsigned short z`                                           |
| 7                 | `placeBlock`                               | 在当前画笔的位置，带着方块数据 `blockData` 放置一个方块，此方块在方块池中的ID为 `blockID`。 | `unsigned short blockID`<br/>`unsigned short blockData`      |
| 8                 | `Z++`                                      | 将画笔的 `Z` 坐标增加 `1`，由于它并不会把什么东西搞没，所以也不弃用了；但这玩意也不再在当下版本的 PhonixBuilder 中被使用了。 | -                                                            |
| 9                 | `NOP`                                      | 摆烂。（不进行操作(No Operation)）                           | -                                                            |
| 10, `0x0A`        | `jumpX`                                    | **(已弃用)** 将画笔的 `X` 坐标增加 `x`，顺带把画笔的 `Y` 和 `Z` 坐标重置为 `0`。 由于命令的实际功能与其名称之间存在差异，该方法在我们生成过程中不会再使用。虽然它已被我们弃用，但在读取 `bdx` 时仍然需要实现它的解析，因为包含此命令的 `bdx` 文件，也就是旧版本的文件一直都在。<br/>而 `jumpX` 与 `addX` 指令之间的差异在于 `jumpX` 的参数用的是 `unsigned int` 而不是 `unsigned short`. | `unsigned int x`                                             |
| 11, `0x0B`        | `jumpY`                                    | **(已弃用)** 将画笔的 `Y` 坐标增加 `y`，顺带把画笔的 `Z` 坐标重置为 `0`。（同上理） | `unsigned int y`                                             |
| 12, `0x0C`        | `jumpZ`                                    | 将画笔的 `Z` 坐标增加 `z`，哦我的天哪，这竟然并没有被弃用，那是因为它并不会把什么东西搞没；可是，理所应当的，这玩意也不再在当下版本的 PhonixBuilder 中被使用了。（同上理） | `unsigned int z`                                             |
| 13, `0x0D`        | `reserved`                                 | 预留命令，你的程序中不应使用此命令                           | -                                                            |
| 14, `0x0E`        | `*X++`                                     | 将画笔的 `X` 坐标增加 `1`。                                  | -                                                            |
| 15, `0x0F`        | `*X--`                                     | 将画笔的 `X` 坐标减少 `1`。                                  | -                                                            |
| 16, `0x10`        | `*Y++`                                     | 将画笔的 `Y` 坐标增加 `1`。                                  | -                                                            |
| 17, `0x11`        | `*Y--`                                     | 将画笔的 `Y` 坐标减少 `1`。                                  | -                                                            |
| 18, `0x12`        | `*Z++`                                     | 将画笔的 `Z` 坐标增加 `1`。                                  | -                                                            |
| 19, `0x13`        | `*Z--`                                     | 将画笔的 `Z` 坐标减少 `1`。                                  | -                                                            |
| 20, `0x14`        | `addX(int16_t)`                            | 将画笔的 `X` 坐标增加 `x`，此 `x` 可为正、为负或为零。       | `short x`                                                    |
| 21, `0x15`        | `addX(int32_t)`                            | 将画笔的 `X` 坐标增加 `x`，此指令与前述（20）之异乃参数之选用：此指令使用 `int32` 为其参数 | `int x`                                                      |
| 22, `0x16`        | `addY(int16_t)`                            | 将画笔的 `Y` 坐标增加 `y`。（同上理）                        | `short y`                                                    |
| 23, `0x17`        | `addY(int32_t)`                            | 将画笔的 `Y` 坐标增加 `y`。（同上理）                        | `int y`                                                      |
| 24, `0x18`        | `addZ(int16_t)`                            | 将画笔的 `Z` 坐标增加 `z`。（同上理）                        | `short z`                                                    |
| 25, `0x19`        | `addZ(int32_t)`                            | 将画笔的 `Z` 坐标增加 `z`。（同上理）                        | `int z`                                                      |
| 26, `0x1A`        | `assignCommandBlockData`                   | **(已弃用, 可以采用 `36` 指令取代)** 在画笔当前位置的方块设置指令方块的数据 *\[译注：这里可能是说，无论是啥方块都可以加指令方块的数据，但是嘞，只有指令方块才能起效\]* | `unsigned int mode {脉冲 = 0, 循环 = 1, 连锁 = 2}` <br/> `char *command` <br/> `char *customName` <br/> `char *lastOutput (此项无效，可被设为 '\0')` <br/> `int tickdelay` <br/> `bool executeOnFirstTick` <br/> `bool trackOutput` <br/> `bool conditional` <br/> `bool needRedstone` |
| 27, `0x1B`        | `placeCommandBlockWithData`                | **(已弃用, 可以采用 `36` 指令取代)** 在当前笔刷的位置放一个命令方块，并设置其数据值。 | `unsigned short blockID` <br/> `unsigned short blockData` <br/> `unsigned int mode {脉冲 = 0, 循环 = 1, 连锁 = 2}` <br/> `char *command` <br/> `char *customName` <br/> `char *lastOutput (此项无效，可被设为 '\0')` <br/> `int tickdelay` <br/> `bool executeOnFirstTick` <br/> `bool trackOutput` <br/> `bool conditional` <br/> `bool needRedstone` |
| 28, `0x1C`        | `addX(int8_t)`                             | 将画笔的 `X` 坐标增加 `x`，此指令与前述（20）之异乃参数之选用：此指令使用 `char` 为其参数 | `char x //int8_t x`                                          |
| 29, `0x1D`        | `addY(int8_t)`                             | 将画笔的 `Y` 坐标增加 `y`。（同上理）                        | `char y //int8_t y`                                          |
| 30, `0x1E`        | `addZ(int8_t)`                             | 将画笔的 `Z` 坐标增加 `z`。（同上理）                        | `char z //int8_t z`                                          |
| 31, `0x1F`        | `useRuntimeIdPalette`                      | 使用预设的运行时ID方块池。`presetId`(预设ID) 是 PhoenixBuilder 内的值。当前网易我的世界基岩版( 1.17.0 @ 2.0.5 )下的 `presetId` 是 `117`。 每一个运行时ID都对应着一个方块，而且包含其数据值。<br/> 详见 [fastbuilder/world_provider/runtimeIds.json](fastbuilder/world_provider/runtimeIds.json) | `unsigned char presetId`                                     |
| 32, `0x20`        | `placeBlockWithRuntimeId(uint16_t)`        | 使用特定的运行时ID在当前画笔的位置放置方块。                 | `unsigned short runtimeId`                                   |
| 33, `0x21`        | `placeBlockWithRuntimeId`                  | 使用特定的运行时ID在当前画笔的位置放置方块。                 | `unsigned int runtimeId`                                     |
| 34, `0x22`        | `placeCommandBlockWithRuntimeId(uint16_t)` | 使用特定的运行时ID在当前画笔的位置放置命令方块，并设置其数据值。 | `unsigned short runtimeId` <br/> `unsigned int mode {脉冲 = 0, 循环 = 1, 连锁 = 2}` <br/> `char *command` <br/> `char *customName` <br/> `char *lastOutput (此项无效，可被设为 '\0')` <br/> `int tickdelay` <br/> `bool executeOnFirstTick` <br/> `bool trackOutput` <br/> `bool conditional` <br/> `bool needRedstone` |
| 35, `0x23`        | `placeCommandBlockWithRuntimeId`           | 使用特定的运行时ID在当前画笔的位置放置指令方块，并设置其数据值。 | `unsigned short runtimeId` <br/> `unsigned int mode {脉冲 = 0, 循环 = 1, 连锁 = 2}` <br/> `char *command` <br/> `char *customName` <br/> `char *lastOutput (此项无效，可被设为 '\0')` <br/> `int tickdelay` <br/> `bool executeOnFirstTick` <br/> `bool trackOutput` <br/> `bool conditional` <br/> `bool needRedstone` |
| 36, `0x24`        | `placeCommandBlockWithDataNew`             | 使用特定的数据值在当前画笔的位置放置指令方块，并设置其数据值。 | `unsigned short data` <br/> `unsigned int mode {脉冲 = 0, 循环 = 1, 连锁 = 2}` <br/> `char *command` <br/> `char *customName` <br/> `char *lastOutput (此项无效，可被设为 '\0')` <br/> `int tickdelay` <br/> `bool executeOnFirstTick` <br/> `bool trackOutput` <br/> `bool conditional` <br/> `bool needRedstone` |
| 37, `0x25`        | `placeBlockWithChestData(uint16_t)`        | 在画笔所在位置放置一个箱子，并加入指定的数据。 `slotCount`(物品栏数) 的数据类型为 `unsigned char`，因为我的世界用一个字节来存储物品栏编号。 | `unsigned short runtimeId` <br/> `unsigned char slotCount` <br/> `struct ChestData data` |
| 38, `0x26`        | `placeBlockWithChestData`                  | 在画笔所在位置放置一个箱子，并载入指定的数据。 `slotCount`(物品栏数) 的数据类型为 `unsigned char`，因为我的世界用一个字节来存储物品栏编号。 | `unsigned int runtimeId`<br/>`unsigned char slotCount`<br/>`struct ChestData data` |
| 88, `'X'`, `0x58` | `end`                                      | 停止读入。注意！虽然通常的结尾应该是 "XE" （2字节），但是用 'X' （1字节）是允许的。 | -                                                            |
| 90, `0x5A`        | `isSigned`                                 | 这是一个与其他命令功能稍有不同的命令，其参数应当出现在其前面，而这个指令呢也只能出现在文件的末尾。在不知道所以然的情况下，请不要使用它，因为无效的签名会使得 PhoenixBuilder 无法去构建你的结构。详见 `签名` 部分。 | `unsigned char signatureSize`                                |

此表为 bdump v4 到 2022/1/29 为止的全部指令。

此外，对于 `struct ChestData` 数据结构，应当如下：

```
struct ChestData {
	char *itemName;
	unsigned char count;
	unsigned short data;
	unsigned char slotID;
}
```

（以下部分目前未被更新，但理应正常运作）

*\[译注：这（指上面那句话）可跟我一点毛线关系都没有啊，原文都是这样写的昂\]*

下面是一些 `bdx` 文件的例子。

假设我们是一个熊孩子，来放置一个TNT在 `{3,5,6}`(**相对坐标**) 上，顺带地再放一个循环指令方块，里面写着 `kill @e[type=tnt]` 还加了悬浮字 `Kill TNT!` ，且始终启用，放在 `{3,6,6}` 上，再顺手一点，我们放一块恶臭的玻璃在 `{114514,15,1919810}` 上，一块恶臭的铁块在 `{114514,15,1919800}` 上。好了，那么未被压缩的 BDX 文件应为如下：

`BDX\0DEMO\0\x01tnt\0\x1C\x03\x01repeating_command_block\0\x01glass\0\x01iron_block\0\x1E\x06\x1D\x05\x07\0\0\0\0\x10\x1B\0\x01\0\0\x01kill @e[type=tnt]\0Kill TNT!\0\0\0\0\0\0\x01\x01\0\0\x1D\x09\x19\0\x1D\x4B\x3C\x15\0\x01\xBF\x4F\x07\0\x02\0\0\x1E\xF6\x07\0\x03\0\0XE`

下面是伪代码形式的指令表达法，便于我们观察此结构具体的运作模式。

```assembly
author 'DEMO\0'
addToBlockPalette 'tnt\0' ; 方块ID: 0
addSmallX 3 ; 画笔位置: {3,0,0}
addToBlockPalette 'repeating_command_block\0' ; 方块ID: 1
addToBlockPalette 'glass\0' ; 方块ID: 2
addToBlockPalette 'iron_block\0' ; 方块ID: 3
addSmallZ 6 ; 画笔位置: {3,0,6}
addSmallY 5 ; 画笔位置: {3,5,6}
placeBlock (int16_t)0, (int16_t)0 ; TNT将会被放在 {3,5,6}
NewYadd ; *Y++, 画笔位置: {3,6,6}
placeCommandBlockWithData (int16_t)1, (int16_t)0, 1, 'kill @e[type=tnt]\0', 'Kill TNT!\0', '\0', (int32_t)0, 1, 1, 0, 0 ; 指令方块将会被放在 {3,6,6}
addSmallY 9 ; 画笔位置: {3,15,6}
addBigZ 1919804 ; 1919810: 00 1D 4B 3C = 01d4b3ch, 画笔位置: {3,15,1919810}
addBigX 114511 ; 114511: 00 01 BF 4F = 01bf4fh, 画笔位置: {114514,15,1919810}
placeBlock (int16_t)2,(int16_t)0 ; 玻璃将会被放在 {114514,15,1919810}
addSmallZ -10 ; -10: F6 = 0f6h, 画笔位置: {114514,15,1919800}
placeBlock (int16_t)3,(int16_t)0 ; 铁块 将会被放在 {114514,15,1919800}
end
db 'E'
```

## 签名

*PhoenixBuilder* 0.3.5 实现了一个 bdump 文件签名系统，用以辨认文件**真正的**发布者。虽然使用 PGP 进行签名是一种良好且安全的方式，但我们选择了一种高度依赖于我们的身份验证服务器的签名方法，因为仅仅为可以随时连接到服务器的在线程序实现PGP签名毫无意义。


请注意， `bdx` 文件可不必被签名，除非用户打开了 `-S`（严格）开关。但这并不妨碍你去给他签名，如果你为了签名而签名的话，则应确保其正常工作，因为 *PhoenixBuilder* 会拒绝处理签名不正确的 `bdx` 文件。

### API

先让我们看看这 `bdx` 文件的签名接口叭。通过以下两个过程，我们就可以轻易签名了。

请使用 `HTTPS` 链接来连接我们接口的主机 `uc.fastbuilder.pro` 。

#### 签名过程

* 发送请求(Request)：

    ```http
    POST /signbdx.web HTTP/1.1
    Host: uc.fastbuilder.pro
    User-Agent: MyApplication/0.1
    
    {"hash": "<未压缩的，且不含结束指令'X'的bdx文件的哈希值>","token": "<你的FastBuilder密钥(Token)>"}
    ```

* 返回应答(Response)：

  ```http
  HTTP/1.1 200 OK
  Content-Type: application/json
  
  {"success":true,"sign":"<签名的Base64值>",message:""}
  ```

#### 验证过程

* 发送请求(Request)：

    ```http
    POST /verifybdx.web HTTP/1.1
    Host: uc.fastbuilder.pro
    User-Agent: MyApplication/0.1
    
    {"hash": "<未压缩的，且不含结束指令'X'的bdx文件的哈希值>","sign": "<签名的Base64值>"}
    ```

* 返回应答(Response)：

  ```http
  HTTP/1.1 200 OK
  Content-Type: application/json
  
  {"success":true,"corrupted":false,"username":"<签名人>",message:""}
  ```

在签名后，签名的 base64 值应在解码后再写入文件已压缩的部分，后面跟着签名长度(1 字节)和`isSigned`标志。
