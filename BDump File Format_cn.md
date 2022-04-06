 - 注意，此中文译本非官方开发者翻译。我也不知道是不是完全正确，如果主文件有更新，请联系我继续翻译。 ——金羿
 - 翻译日期 2022 4 5， 原文日期 2022 1 29

# BDump文件结构

BDump v3 是个用于存储*我的世界*建筑结构的文件格式。其由多组描述建造过程的指令而组成。

按照一定的顺序来写下每一个方块的ID来存入文件是一个不错的方案，同时也能够减小文件大小，但是这样的方案会包含了很多不需要的空气方块，而他们又会增大这~~可怜的~~文件大小。因此我们设计了一种新的文件格式，用一个指针来指向“画笔”应该放在哪里，这种文件格式有一系列的指令来告诉这个“画笔”应该怎么样移动、在哪里去放置什么方块。使用这种文件格式，空气方块可以非常好地避免，所以我们的文件也会更为轻便。

## 基础文件结构

BDump v3 文件的后缀名为`.bdx`，且文件头为`BD@`, 这就表明此类文件使用的压缩算法为 brotli（官方压缩质量为6）。请注意，也有一种文件头为`BDZ`的BDump文件，这表明此文件使用的压缩算法为gzip算法，然而这种文件头所代表的`.bdx`文件很快将不被FastBuilder Phoenix支持，因为它本身就被反对了蛮久并且也难以再找到此类型的文件了。我们将这种文件头定义为“压缩头”(compression header)，并且在此压缩头下面的内容将以压缩头所表明的方式进行压缩。

> 小注: BDump v2 的文件后缀是 `.bdp`，且文件头为 `BDMPS\0\x02\0`。

在压缩头之后的，压缩后内容的起始字符为 `BDX\0`，且作者的游戏名紧跟其后，并以 `\0` 表示其玩家名的表示完毕。*\[译注：即若作者之游戏名为Eilles，则此文件压缩后的内容应以*`BDX\0Eilles\0`*开头\]*此后之文本即含参指令了，它们惜惜相依，紧紧相连。每个指令的ID占有1字节(byte)的空间，其正是`无符号单字`(`unsigned char`)所占空间。

所有的操作都基于一个用以标识“画笔”所在位置的 `Vec3` 值。

来来来，我们看看指令列表先。

> 请注意：整型的数据将会以<font style="color:red;">**大端字节序**</font>(big endian)记录.
>
> 大小端字节序有何不同呢？
>
> *\[译注：你完全可以去查百度、必应上面搜索出来的解析，那玩意肯定让你半蒙半懂，但这玩意本身相对而言也并非十分绝对得重要，你看下面这个全蒙的也挺好。\]*
>
> 例如，一个`int32`的`1`在小端字节序的表示下，内存中是这样的`01 00 00 00`，而大端为`00 00 00 01`。

类型定义如下：

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
* `布尔`(bool)：1字节长的布尔(亦称逻辑)数据，仅可为真(true, 1)或假(false, 0)



# ====美好的事物总是后一步到来 TO BE TRANSLATED OF THIS TABLE====

| ID                | 内部名                                      | 描述                                                    | 参数                                                    |
| ----------------- | ------------------------------------------ | ------------------------------------------------------------ | ------------------------------------------------------------ |
| 1                 | `addToBlockPalette`                        | Add a specific block name to the palette, and the id of the block is sorted in the term of the command called, e.g. the id of the first time calling the command is `0`, and the id of the second time is `1`. The maximum number of types of blocks is `65536`. | `char *blockName`                                            |
| 2                 | `addX`                                     | **(DEPRECATED)** Add `x` to the brush position's `X`, and reset the value of `Y` and `Z` to `0`. This method is deprecated since the difference between the real function of the command and what it's name say it should do. Though it's deprecated, you still need to implement it's parsing since there's still `bdx` files containing this command. | `unsigned short x`                                           |
| 3                 | `X++`                                      | **(DEPRECATED) **Add `1` to the brush position's `X`, and reset the value of `Y` and `Z` to `0`. | -                                                            |
| 4                 | `addY`                                     | **(DEPRECATED)** Add `y` to the brush position's `Y`, and reset the value of `Z` to `0`. | `unsigned short y`                                           |
| 5                 | `Y++`                                      | **(DEPRECATED) **Add `1` to the brush position's `Y`, and reset the value of `Z` to `0`. | -                                                            |
| 6                 | `addZ`                                     | Add `z` to the brush position's `Z`, it's not deprecated since it resets nothing, though it's no longer used by the current version of PhonixBuilder. | `unsigned short z`                                           |
| 7                 | `placeBlock`                               | Place a block on the current position of the brush with the id from the `addToBlockPalette` command and the `blockData` value. | `unsigned short blockID`<br/>`unsigned short blockData`      |
| 8                 | `Z++`                                      | Add `1` to the brush position's `Z`, it's not deprecated since it resets nothing, though it's no longer used by the current version of PhonixBuilder. | -                                                            |
| 9                 | `NOP`                                      | Do nothing. (No Operation)                                   | -                                                            |
| 10, `0x0A`        | `jumpX`                                    | **(DEPRECATED)** Add `x` to the brush position's `X`, and reset the value of `Y` and `Z` to `0`. This method is deprecated since the difference between the real function of the command and what it's name say it should do. Though it's deprecated, you still need to implement it's read since there's still bdx files contain this command.<br/>The difference between `jumpX` and `addX` command is that `jumpX` uses `unsigned int` for its argument instead of `unsigned short`. | `unsigned int x`                                             |
| 11, `0x0B`        | `jumpY`                                    | **(DEPRECATED)** Add `y` to the brush position's `Y`, and reset the value of `Z` to `0`. | `unsigned int y`                                             |
| 12, `0x0C`        | `jumpZ`                                    | Add `z` to the brush position's `Z`, it's not deprecated since it resets nothing, though it's no longer used by the current version of PhonixBuilder. | `unsigned int z`                                             |
| 13, `0x0D`        | `reserved`                                 | Reserved command, shouldn't be used by your program.         | -                                                            |
| 14, `0x0E`        | `*X++`                                     | Add `1` to the brush position's `X`.                         | -                                                            |
| 15, `0x0F`        | `*X--`                                     | Subtract `1` from the brush position's `X`.                  | -                                                            |
| 16, `0x10`        | `*Y++`                                     | Add `1` to the brush position's `Y`.                         | -                                                            |
| 17, `0x11`        | `*Y--`                                     | Subtract `1` from the brush position's `Y`.                  | -                                                            |
| 18, `0x12`        | `*Z++`                                     | Add `1` to the brush position's `Z`.                         | -                                                            |
| 19, `0x13`        | `*Z--`                                     | Subtract `1` from the brush position's `Z`.                  | -                                                            |
| 20, `0x14`        | `addX(int16_t)`                            | Add `x` to the brush position's `X`. `x` could be either positive, negative or zero. | `short x`                                                    |
| 21, `0x15`        | `addX(int32_t)`                            | Add `x` to the brush position's `X`. The difference between this command and the previous one is this command uses `int32` as its argument. | `int x`                                                      |
| 22, `0x16`        | `addY(int16_t)`                            | Add `y` to the brush position's `Y`.                         | `short y`                                                    |
| 23, `0x17`        | `addY(int32_t)`                            | Add `y` to the brush position's `Y`.                         | `int y`                                                      |
| 24, `0x18`        | `addZ(int16_t)`                            | Add `z` to the brush position's `Z`.                         | `short z`                                                    |
| 25, `0x19`        | `addZ(int32_t)`                            | Add `z` to the brush position's `Z`.                         | `int z`                                                      |
| 26, `0x1A`        | `assignCommandBlockData`                   | Set the command block data for the block at the brush's position.**(DEPRECATED, USE COMMAND 36 INSTEAD)** | `unsigned int mode {Impulse=0, Repeat=1, Chain=2}`<br/>`char *command`<br/>`char *customName`<br/>`char *lastOutput (no effect and can be set to'\0')`<br/>`int tickdelay`<br/>`bool executeOnFirstTick`<br/>`bool trackOutput`<br/>`bool conditional`<br/>`bool needRedstone` |
| 27, `0x1B`        | `placeCommandBlockWithData`                | Place a command block, and set its data at the brush's position.**(DEPRECATED,USE COMMAND 36 INSTEAD)** | `unsigned short blockID`<br/>`unsigned short blockData`<br/>`unsigned int mode {Impulse=0, Repeat=1, Chain=2}`<br/>`char *command`<br/>`char *customName`<br/>`char *lastOutput (no effect and can be set to'\0')`<br/>`int tickdelay`<br/>`bool executeOnFirstTick`<br/>`bool trackOutput`<br/>`bool conditional`<br/>`bool needRedstone` |
| 28, `0x1C`        | `addX(int8_t)`                             | Add `x` to the brush position's `X`. The difference between this command and the `*addX` command is that this command uses `char` as its argument. | `char x //int8_t x`                                          |
| 29, `0x1D`        | `addY(int8_t)`                             | Add `y` to the brush position's `Y`.                         | `char y //int8_t y`                                          |
| 30, `0x1E`        | `addZ(int8_t)`                             | Add `z` to the brush position's `Z`.                         | `char z //int8_t z`                                          |
| 31, `0x1F`        | `useRuntimeIdPalette`                      | Use a preset runtime id palette. `presetId` is the id of the runtime id palette used, which is assigned by PhoenixBuilder itself. The `presetId` for the current version of NetEase's Minecraft BE (1.17.0 @ 2.0.5) is `117`. Each runtime id matches a individual block state (contains its data value)<br/>See [fastbuilder/world_provider/runtimeIds.json](fastbuilder/world_provider/runtimeIds.json) for detailed content. | `unsigned char presetId`                                     |
| 32, `0x20`        | `placeBlockWithRuntimeId(uint16_t)`        | Place a block with a specific runtime id at the brush's position. | `unsigned short runtimeId`                                   |
| 33, `0x21`        | `placeBlockWithRuntimeId`                  | Place a block with a specific runtime id at the brush's position. | `unsigned int runtimeId`                                     |
| 34, `0x22`        | `placeCommandBlockWithRuntimeId(uint16_t)` | Place a command block with the specified runtime id, and set its data at the brush's position. | `unsigned short runtimeId`<br/>`unsigned int mode {Impulse=0, Repeat=1, Chain=2}`<br/>`char *command`<br/>`char *customName`<br/>`char *lastOutput (no effect and can be set to'\0')`<br/>`int tickdelay`<br/>`bool executeOnFirstTick`<br/>`bool trackOutput`<br/>`bool conditional`<br/>`bool needRedstone` |
| 35, `0x23`        | `placeCommandBlockWithRuntimeId`           | Place a command block with the specified runtime id, and set its data at the brush's position. | `unsigned int runtimeId`<br/>`unsigned int mode {Impulse=0, Repeat=1, Chain=2}`<br/>`char *command`<br/>`char *customName`<br/>`char *lastOutput (no effect and can be set to'\0')`<br/>`int tickdelay`<br/>`bool executeOnFirstTick`<br/>`bool trackOutput`<br/>`bool conditional`<br/>`bool needRedstone` |
| 36, `0x24`        | `placeCommandBlockWithDataNew`             | Place a command block with the specified data value, and set its data at the brush's position. | `unsigned short data`<br/>`unsigned int mode {Impulse=0, Repeat=1, Chain=2}`<br/>`char *command`<br/>`char *customName`<br/>`char *lastOutput (no effect and can be set to'\0')`<br/>`int tickdelay`<br/>`bool executeOnFirstTick`<br/>`bool trackOutput`<br/>`bool conditional`<br/>`bool needRedstone` |
| 37, `0x25`        | `placeBlockWithChestData(uint16_t)`        | Place a (chest) block with the specified data within the chest. `slotCount`'s type is `unsigned char` since Minecraft uses only a byte for the slot ID. | `unsigned short runtimeId`<br/>`unsigned char slotCount`<br/>`struct ChestData data` |
| 38, `0x26`        | `placeBlockWithChestData`                  | Place a (chest) block with the specified data within the chest. `slotCount`'s type is `unsigned char` since Minecraft uses only a byte for the slot ID. | `unsigned int runtimeId`<br/>`unsigned char slotCount`<br/>`struct ChestData data` |
| 88, `'X'`, `0x58` | `end`                                      | Stop reading. Note that though the general end is "XE" (2 bytes long), but a 'X' (1 byte long) character is enough. | -                                                            |
| 90, `0x5A`        | `isSigned`                                 | A command that functions a little different with other commands, its argument is the previous byte of it, would only appear in the end of the file. Please do not use it unless you know how to use since an invalid signature would prevent PhoenixBuilder from constructing your structure. See paragraph `Signing` for details. | `unsigned char signatureSize`                                |

此表为 bdump v4 到 2022/1/29 为止的全部指令。

对于 `struct ChestData` 数据结构，应当如下：

```
struct ChestData {
	char *itemName;
	unsigned char count;
	unsigned short data;
	unsigned char slotID;
}
```

（以下部分目前未被更新，但理应正常运作）

*\[译注：这可跟我一点毛线关系都没有啊，原文都是这样写的昂\]*

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

*FastBuilder Phoenix* 0.3.5 实现了一个 bdump 文件签名系统，用以辨认文件**真正的**发布者。虽然使用 PGP 进行签名是一种良好且安全的方式，但我们选择了一种高度依赖于我们的身份验证服务器的签名方法，因为仅仅为可以随时连接到服务器的在线程序实现PGP签名毫无意义。


请注意， `bdx` 文件可不必被签名，除非用户打开了 `-S`（严格）开关。但这并不妨碍你去给他签名，如果你为了签名而签名的话，则应确保其正常工作，因为签名不正确的 `bdx` 文件是无法被 *FastBuilder Phoenix* 处理的。

### 应用程序接口(API)

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

在请求签名接口之后，签名的 base64 值应该被解压缩并写入文件的已压缩部分，接着是签名长度和`isSigned`标志。

*\[我不理解……我不知道这句话怎么翻译……所以我就……直译了；各位，给你们看一下原话： *`After requesting the signing api, the base64 value of the signature should be decompressed and written to the compressed part of file, then the signature length and isSigned flag.`* \]*
