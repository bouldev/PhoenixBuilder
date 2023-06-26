# BDump文件格式

> 感谢 [EillesWan](https://github.com/EillesWan) 提供的翻译，本文档内容主要来自其翻译内容，存在部分改动。
>
> 部分注有 `译注` 的内容均为此贡献者所注。


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
| 1                 | `CreateConstantString`                     | 将特定的 `字符串` 放入 `方块池` 。`字符串` 在 `方块池` 中的 `ID` 将按照调用此命令的顺序进行排序。如：你第一次调用这个命令的时候，对应 `字符串` 的 `ID` 为 `0` ，第二次就是 `1` 了。你最多只能添加到 `65535`<br/>*\[译注：通常情况下，`字符串` 是一个方块的 `英文ID名` ，如 `glass` \]* | `char *constantString` |
| 2                 | **已弃用且已移除**                          | - | - |
| 3                 | **已弃用且已移除**                          | - | - |
| 4                 | **已弃用且已移除**                          | - | - |
| 5                 | `PlaceBlockWithBlockStates`                | 在画笔所在位置放置一个方块。同时指定欲放置方块的 `方块状态` 在方块池中的 `ID` 为 `blockStatesConstantStringID` ，且该方块在方块池中的 `ID` 为 `blockConstantStringID`<br/> `方块状态` 的格式形如 `["color": "orange"]` | `unsigned short blockConstantStringID`<br/>`unsigned short blockStatesConstantStringID` |
| 6                 | `AddInt16ZValue0`                          | 将画笔的 `Z` 坐标增加 `value` | `unsigned short value` |
| 7                 | `PlaceBlock`                               | 在画笔所在位置放置一个方块。同时指定欲放置方块的 `数据值(附加值)` 为 `blockData` ，且该方块在方块池中的 `ID` 为 `blockConstantStringID` | `unsigned short blockConstantStringID`<br/>`unsigned short blockData` |
| 8                 | `AddZValue0`                               | 将画笔的 `Z` 坐标增加 `1` | - |
| 9                 | `NOP`                                      | 摆烂，即不进行操作(`No Operation`) | - |
| 10, `0x0A`        | **已弃用且已移除**                          | - | - |
| 11, `0x0B`        | **已弃用且已移除**                          | - | - |
| 12, `0x0C`        | `AddInt32ZValue0`                          | 将画笔的 `Z` 坐标增加 `value` | `unsigned int value` |
| 13, `0x0D`        | `PlaceBlockWithBlockStates`                | 在画笔所在位置放置一个方块。同时指定欲放置方块的 `方块状态` 为 `blockStatesString` ，且该方块在方块池中的 `ID` 为 `blockConstantStringID`<br/> `方块状态` 的格式形如 `["color": "orange"]` | `unsigned short blockConstantStringID`<br/>`char *blockStatesString` |
| 14, `0x0E`        | `AddXValue`                                | 将画笔的 `X` 坐标增加 `1` | - |
| 15, `0x0F`        | `SubtractXValue`                           | 将画笔的 `X` 坐标减少 `1` | - |
| 16, `0x10`        | `AddYValue`                                | 将画笔的 `Y` 坐标增加 `1` | - |
| 17, `0x11`        | `SubtractYValue`                           | 将画笔的 `Y` 坐标减少 `1` | - |
| 18, `0x12`        | `AddZValue`                                | 将画笔的 `Z` 坐标增加 `1` | - |
| 19, `0x13`        | `SubtractZValue`                           | 将画笔的 `Z` 坐标减少 `1` | - |
| 20, `0x14`        | `AddInt16XValue`                           | 将画笔的 `X` 坐标增加 `value` 且 `value` 可正可负，亦或 `0` | `short value` |
| 21, `0x15`        | `AddInt32XValue`                           | 将画笔的 `X` 坐标增加 `value`<br/>此指令与上一命令的不同点是此指令使用 `int32_t` 作为其参数 | `int value` |
| 22, `0x16`        | `AddInt16YValue`                           | 将画笔的 `Y` 坐标增加 `value` （同上理） | `short value` |
| 23, `0x17`        | `AddInt32YValue`                           | 将画笔的 `Y` 坐标增加 `value` （同上理） | `int value` |
| 24, `0x18`        | `AddInt16ZValue`                           | 将画笔的 `Z` 坐标增加 `value` （同上理） | `short value` |
| 25, `0x19`        | `AddInt32ZValue`                           | 将画笔的 `Z` 坐标增加 `value` （同上理） | `int value` |
| 26, `0x1A`        | `SetCommandBlockData`                      | **(推荐使用 `36` 号命令)** 在画笔当前位置的方块设置指令方块的数据 *\[译注：这里可能是说，无论是啥方块都可以加指令方块的数据，但是嘞，只有指令方块才能起效\]* | `unsigned int mode {脉冲=0, 重复=1, 连锁=2}`<br/>`char *command`<br/>`char *customName`<br/>`char *lastOutput (此项无效，可被设为 '\0')`<br/>`int tickdelay`<br/>`bool executeOnFirstTick`<br/>`bool trackOutput`<br/>`bool conditional`<br/>`bool needsRedstone` |
| 27, `0x1B`        | `PlaceBlockWithCommandBlockData`           | **(推荐使用 `36` 号命令)** 在画笔当前位置放置方块池中 `ID` 为 `blockConstantStringID` 的方块，且该方块的 `方块数据值(附加值)` 为 `blockData` 。放置完成后，为这个方块设置 `命令方块` 的数据(若可行的话) | `unsigned short blockConstantStringID`<br/>`unsigned short blockData`<br/>`unsigned int mode {脉冲=0, 重复=1, 连锁=2}`<br/>`char *command`<br/>`char *customName`<br/>`char *lastOutput (此项无效，可被设为 '\0')`<br/>`int tickdelay`<br/>`bool executeOnFirstTick`<br/>`bool trackOutput`<br/>`bool conditional`<br/>`bool needRedstone` |
| 28, `0x1C`        | `AddInt8XValue`                            | 将画笔的 `X` 坐标增加 `value`<br/>此指令与命令 `AddInt16XValue(20) `的不同点是此指令使用 `char` 作为其参数 | `char value //int8_t value` |
| 29, `0x1D`        | `AddInt8YValue`                            | 将画笔的 `Y` 坐标增加 `value` （同上理） | `char value //int8_t value` |
| 30, `0x1E`        | `AddInt8ZValue`                            | 将画笔的 `Z` 坐标增加 `value` （同上理） | `char value //int8_t value` |
| 31, `0x1F`        | `UseRuntimeIDPool`                         | 使用预设的 `运行时ID方块池`<br/>`poolId`(预设ID) 是 PhoenixBuilder 内的值。网易MC( 1.17.0 @ 2.0.5 )下的 `poolId` 被我们定为 `117`。 每一个 `运行时ID` 都对应着一个方块，而且包含其 `方块数据值(附加值)`<br/>相关内容详见 [PhoenixBuilder/resources](https://github.com/LNSSPsd/PhoenixBuilder/tree/main/resources)<br/>**已不再在新版本中被使用** | `unsigned char poolId` |
| 32, `0x20`        | `PlaceRuntimeBlock`                        | 使用特定的 `运行时ID` 在当前画笔的位置放置方块 | `unsigned short runtimeId`                                   |
| 33, `0x21`        | `placeBlockWithRuntimeId`                  | 使用特定的 `运行时ID` 在当前画笔的位置放置方块 | `unsigned int runtimeId`                                     |
| 34, `0x22`        | `PlaceRuntimeBlockWithCommandBlockData`    | 使用特定的 `运行时ID` 在当前画笔的位置放置命令方块，并设置其数据 | `unsigned short runtimeId`<br/>`unsigned int mode {脉冲=0, 重复=1, 连锁=2}`<br/>`char *command`<br/>`char *customName`<br/>`char *lastOutput (此项无效，可被设为 '\0')`<br/>`int tickdelay`<br/>`bool executeOnFirstTick`<br/>`bool trackOutput`<br/>`bool conditional`<br/>`bool needRedstone` |
| 35, `0x23`        | `PlaceRuntimeBlockWithCommandBlockDataAndUint32RuntimeID` | 使用特定的 `运行时ID` 在当前画笔的位置放置指令方块，并设置其数据 | `unsigned int runtimeId`<br/>`unsigned int mode {脉冲 = 0, 循环 = 1, 连锁 = 2}`<br/>`char *command`<br/>`char *customName`<br/>`char *lastOutput (此项无效，可被设为 '\0')`<br/>`int tickdelay`<br/>`bool executeOnFirstTick`<br/>`bool trackOutput`<br/>`bool conditional`<br/>`bool needRedstone` |
| 36, `0x24`        | `PlaceCommandBlockWithCommandBlockData`    | 根据给定的 `方块数据值(附加值)` 在当前画笔所在位置放置一个指令方块，并设置其数据值 | `unsigned short data`<br/>`unsigned int mode {脉冲 = 0, 循环 = 1, 连锁 = 2}`<br/>`char *command`<br/>`char *customName`<br/>`char *lastOutput (此项无效，可被设为 '\0')`<br/>`int tickdelay`<br/>`bool executeOnFirstTick`<br/>`bool trackOutput`<br/>`bool conditional`<br/>`bool needRedstone` |
| 37, `0x25`        | `PlaceRuntimeBlockWithChestData`           | 在画笔所在位置放置一个 `runtimeId`(特定的 `运行时ID`) 所表示的方块，并向此方块载入数据<br/>其中 `slotCount` 的数据类型为 `unsigned char`，因为我的世界用一个字节来存储物品栏编号。此参数指的是要载入的次数，即要载入的 `ChestData` 结构体数量 | `unsigned short runtimeId` <br/> `unsigned char slotCount` <br/> `struct ChestData data` |
| 38, `0x26`        | `PlaceRuntimeBlockWithChestDataAndUint32RuntimeID` | 在画笔所在位置放置一个 `runtimeId`(特定的 `运行时ID`) 所表示的方块，并向此方块载入数据<br/>其中 `slotCount` 的数据类型为 `unsigned char`，因为我的世界用一个字节来存储物品栏编号。此参数指的是要载入的次数，即要载入的 `ChestData` 结构体数量 | `unsigned int runtimeId`<br/>`unsigned char slotCount`<br/>`struct ChestData data` |
| 39, `0x27`        | `AssignDebugData`                          | 记录调试数据，不对建造过程产生任何影响。 | `uint32_t length`<br>`unsigned char buffer[length]` |
| 40, `0x28`        | `PlaceBlockWithChestData`                  | 放置一个 `blockConstantStringID` 所表示的方块，并指定容器数据。 | `uint16_t blockConstantStringID`<br/>`uint16_t blockData`<br/>`struct ChestData data` |
| 41, `0x29`        | `PlaceBlockWithNBTData`                    | 放置一个 `blockConstantStringID` 所表示的方块且指定它的 `方块状态` 在方块池中的 `ID` 为 `blockStatesConstantStringID`，然后指定 `void *buffer` 所表示的由小端序 NBT 所存储的 `方块实体` 数据<br/>因为一些失误，`blockStatesConstantStringID` 会被记录两次 | `uint16_t blockConstantStringID`<br/>`uint16_t blockStatesConstantStringID`<br/>`uint16_t blockStatesConstantStringID`<br/>`void *buffer` |
| 88, `'X'`, `0x58` | `Terminate`                                | 停止读入。虽然通常的结尾应该是 `XE` （2字节），但是用 `X` （1字节）是允许的 | - |
| 90, `0x5A`        | `isSigned` (伪命令)                         | 这是一个与其他命令功能稍有不同的命令，其参数应当出现在其前面，而这个指令呢也只能出现在文件的末尾。在不知道所以然的情况下，请不要使用它，因为无效的签名会使得 `PhoenixBuilder` 无法去构建你的结构。详见 `签名` 部分。 | `unsigned char signatureSize` |

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


（下述内容的其中一部分目前未被更新，除去部分已经弃用的命令外，其余应当正常运作）










## 文件样例
下面是一些 `bdx` 文件的例子。
***

假设我们是一个熊孩子，来放置一个TNT在 `{3,5,6}`(**相对坐标**) 上，顺带地再放一个循环指令方块，里面写着 `kill @e[type=tnt]` 还加了悬浮字 `Kill TNT!` ，且始终启用，放在 `{3,6,6}` 上，再顺手一点，我们放一块恶臭的玻璃在 `{114514,15,1919810}` 上，一块恶臭的铁块在 `{114514,15,1919800}` 上。好了，那么未被压缩的 BDX 文件应为如下：

`BDX\0DEMO\0\x01tnt\0\x1C\x03\x01repeating_command_block\0\x01glass\0\x01iron_block\0\x1E\x06\x1D\x05\x07\0\0\0\0\x10\x1B\0\x01\0\0\x01kill @e[type=tnt]\0Kill TNT!\0\0\0\0\0\0\x01\x01\0\0\x1D\x09\x19\0\x1D\x4B\x3C\x15\0\x01\xBF\x4F\x07\0\x02\0\0\x1E\xF6\x07\0\x03\0\0XE`

下面是伪代码形式的指令表达法，便于我们观察此结构具体的运作模式。

```assembly
author 'DEMO\0'
CreateConstantString 'tnt\0' ; 方块ID: 0
AddInt8XValue 3 ; 画笔位置: {3,0,0}
CreateConstantString 'repeating_command_block\0' ; 方块ID: 1
CreateConstantString 'glass\0' ; 方块ID: 2
CreateConstantString 'iron_block\0' ; 方块ID: 3
AddInt8ZValue 6 ; 画笔位置: {3,0,6}
AddInt8YValue 5 ; 画笔位置: {3,5,6}
PlaceBlock (int16_t)0, (int16_t)0 ; TNT将会被放在 {3,5,6}
AddYValue ; *Y++, 画笔位置: {3,6,6}
PlaceCommandBlockWithCommandBlockData (int16_t)1, (int16_t)0, 1, 'kill @e[type=tnt]\0', 'Kill TNT!\0', '\0', (int32_t)0, 1, 1, 0, 0 ; 指令方块将会被放在 {3,6,6}
AddInt8YValue 9 ; 画笔位置: {3,15,6}
AddInt32ZValue 1919804 ; 1919810: 00 1D 4B 3C = 01d4b3ch, 画笔位置: {3,15,1919810}
AddInt32XValue 114511 ; 114511: 00 01 BF 4F = 01bf4fh, 画笔位置: {114514,15,1919810}
PlaceBlock (int16_t)2,(int16_t)0 ; 玻璃将会被放在 {114514,15,1919810}
AddInt8ZValue -10 ; -10: F6 = 0f6h, 画笔位置: {114514,15,1919800}
PlaceBlock (int16_t)3,(int16_t)0 ; 铁块 将会被放在 {114514,15,1919800}
Terminate
db 'E'
```
***
如果希望在画笔所在位置放置一个 `正在燃烧的熔炉` ，且这个 `正在燃烧的熔炉` 的第一格和第三格分别是 `苹果 * 3` 和 `钻石 * 64` ，则那么未被压缩的 BDX 文件应为如下：

`BDX\x00DEMO\x00\x1f\x75\x26\x00\x00\x15\x2c\x02apple\x00\x03\x00\x00\x00diamond\x00\x40\x00\x00\x02XE`

下面是伪代码形式的指令表达法，便于我们观察此结构具体的运作模式。

```assembly
author 'DEMO\0' ; 设置作者为 'DEMO'
UseRuntimeIDPool (unsigned char)117 ; 117: 75
PlaceRuntimeBlockWithChestDataAndUint32RuntimeID (unsigned int)5420, (unsigned char)2 , 'apple\x00', (unsigned char)3, (unsigned short)0, (unsigned char)0, 'diamond\x00', (unsigned char)64, (unsigned short)0, (unsigned char)2
Terminate
db 'E'
```

以下是关于上述用到的 `PlaceRuntimeBlockWithChestDataAndUint32RuntimeID` 的相关解析。<br>
|参数|解释|代码片段|其他/备注|
|-|-|-|-|
|`PlaceRuntimeBlockWithChestDataAndUint32RuntimeID (unsigned int)5420`|在画笔所在位置放置一个 `正在燃烧的熔炉`<br/>因为 `正在燃烧的熔炉` 在 `ID` 为 `117` 的 `运行时ID方块池` 中的 `ID` 是 `5420` |`\x26\x00\x00\x15\x2c`|`5420` 在 `16` 进制下，其 `大端字节序` 表达为 `\x00\x00\x15\x2c`<br/>`unsigned int` 是 `正整数型` ，因此有 `4` 个字节|
|`(unsigned char)2`|向 `正在燃烧的熔炉` 载入 `2` 次数据(载入 `2` 个 `ChestData` 结构体)|`\x02`|`2` 在 `16` 进制下，其 `大端字节序` 表达为 `\x02`<br/>`unsigned char` 是 `无符号字节型` ，因此有 `1` 个字节|
|`apple\x00`|放入 `苹果` |`apple\x00`|`char *` 是以 `\x00`(`UTF-8` 编码)结尾的字符串|
|`(unsigned char)3`|`苹果` 的数量为 `3`|`\x03`|`3` 在 `16` 进制下，其 `大端字节序` 表达为 `\x03`<br/>`unsigned char` 是 `无符号字节型` ，因此有 `1` 个字节|
|`(unsigned short)0`|`苹果` 的 `物品数据值` 为 `0`|`\x00\x00`|`0` 在 `16` 进制下，其 `大端字节序` 表达为 `\x00\x00`<br/>`unsigned short` 是 `无符号短整型` ，因此有 `2` 个字节|
|`(unsigned char)0`|将 `苹果` 放在第 `1` 个槽位|`\x00`|`0` 在 `16` 进制下，其 `大端字节序` 表达为 `\x00`<br/>`unsigned char` 是 `无符号字节型` ，因此有 `1` 个字节<br/>第一个槽位一般使用 `0` ，第二个槽位则为 `1` ，第三个槽位则为 `2` ，以此类推。|
|`diamond\x00`|放入 `钻石`|`diamond\x00`|`char *` 是以 `\x00`(`UTF-8` 编码)结尾的字符串|
|`(unsigned char)64`|`钻石` 的数量为 `64`|`\x40`|`64` 在 `16` 进制下，其 `大端字节序` 表达为 `\x40`<br/>`unsigned char` 是 `无符号字节型` ，因此有 `1` 个字节|
|`(unsigned short)0`|`钻石` 的 `物品数据值` 为 `0`|`\x00\x00`|`0` 在 `16` 进制下，其 `大端字节序` 表达为 `\x00\x00`<br/>`unsigned short` 是 `无符号短整型` ，因此有 `2` 个字节|
|`(unsigned char)2`|将 `钻石` 放在第 `3` 个槽位|`\x02`|`2` 在 `16` 进制下，其 `大端字节序` 表达为 `\x02`<br/>`unsigned char` 是 `无符号字节型` ，因此有 `1` 个字节<br/>第一个槽位一般使用 `0` ，第二个槽位则为 `1` ，第三个槽位则为 `2` ，以此类推。|

您可以在 [PhoenixBuilder/resources](https://github.com/LNSSPsd/PhoenixBuilder/tree/main/resources) 查看 `运行时ID方块池` 。<br>
本样例采用的是 [PhoenixBuilder/resources/blockRuntimeIDs/netease/runtimeIds_117.json](https://github.com/LNSSPsd/PhoenixBuilder/blob/main/resources/blockRuntimeIDs/netease/runtimeIds_117.json) 所述之版本。










## 签名
*PhoenixBuilder* 的 `0.3.5` 版本实现了一个 `bdump 文件签名系统` ，用以辨认文件**真正的**发布者。

请注意， `bdx` 文件可不必被签名，除非用户打开了 `-S`（严格）开关。但这并不妨碍你去给他签名，如果你为了签名而签名的话，则应确保其正常工作，因为 *PhoenixBuilder* 会拒绝处理签名不正确的 `bdx` 文件。

我们使用基于 `RSA` 的哈希方法对 `BDX` 文件进行 `签名` 。签名时，相应的服务器会为每个用户颁发一个单独的认证集，然后 *PhoenixBuilder* 用相应的 `私钥` 对文件进行 `签名` ，并向对应的硬编码服务器提供文件中根密钥链接的 `公钥` ，用于校验 `BDX` 文件的真实发布者。

有关 `签名` 的更多信息及详细细节，另见 `fastbuilder/bdump/utils.go` : `SignBDXNew`/`VerifyBDXNew`
