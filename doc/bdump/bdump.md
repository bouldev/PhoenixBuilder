# BDump File Format

BDump v3 is a file format for storing Minecraft's structures. It is made of different commands that indicate the constructing process.

By writing the ids that represent each blocks in a specific order is a workable plan that reduces the file size, but this would allow a large amount of unexpected air blocks that increasing the file size so we implemented a new format which has a pointer that indicates where the "brush" is, the file is a set of commands that tells the "brush" how to move, and where to place blocks. Within this file format, air blocks can be simply skipped by a move command so files can be smaller.

## Basic File Structure

BDump v3 file's extension is `.bdx`, and the general header of it is `BD@`, which stands for that the file was compressed with brotli compression algorithm (the compress quality phoenixbuilder uses is 6). Note that there's also a header `BDZ` that stands for the file was compressed with gzip compression algorithm, which is no longer supported by PhoenixBuilder today since it has been deprecated for a long time and it's hard to find this type's file again. We define such kind of header as "compression header"  and the content after it is compressed with the compression algorithm it indicates.

> Tip: BDump v2's extension is `.bdp` and the header is `BDMPS\0\x02\0`.

The header of the compressed content is `BDX\0`, and the author's player name that terminated with `\0` is followed right after it. Then the content after it is the command with arguments that written one-by-one tightly. Each command id would take 1 byte of space, like what an `unsigned char` do.

All the operations depend a `Vec3` value that represents the current position of the "brush".

Let's see the list of commands first.

> Note: Integers would be written in <font style="color:red;">**big endian**</font>.
>
> What is the difference of little endian and big endian?
>
> For example, an int32 number in little endian, `1`, is `01 00 00 00` in the memory, and the memory of an int32 number `1` in big endian is `00 00 00 01`.

Type definition:

* {int}: a number that can be positive, negative or zero.
* {unsigned int}: a number that can be positive or zero.
* `char`: an {int} value with 1 byte long.
* `unsigned char`: an {unsigned int} value with 1 byte long.
* `short`: an {int} value with 2 bytes long.
* `unsigned short`: an {unsigned int} value with 2 bytes long.
* `int32_t`: an {int} value with 4 bytes long.
* `uint32_t`: an {unsigned int} value with 4 bytes long.
* `char *`: a string that terminated with `\0` (encoding is utf-8).
* `int`: alias of `int32_t`
* `unsigned int`: alias of `uint32_t`
* `bool`: a value that can be either `true(1)` or `false(0)`, 1 byte long.

| ID                | Internal name                              | Description                                                  | Arguments                                                    |
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

The list above is all the commands of the bdump v4 till 2022-1-29.

For the `struct ChestData` data format:

```
struct ChestData {
	char *itemName;
	unsigned char count;
	unsigned short data;
	unsigned char slotID;
}
```

(Contents below are not updated currently, but they should work correctly.)

Let's see how to make a `bdx` file using these commands.

If we want to place a TNT block at `{3,5,6}`(**relative**), and a repeating command block with command `kill @e[type=tnt]` and name `Kill TNT!` that doesn't need redstone to be activated at `{3,6,6}`, then a glass block at `{114514,15,1919810}` and a iron block at `{114514,15,1919800}`, the uncompressed bdx file might be:

`BDX\0DEMO\0\x01tnt\0\x1C\x03\x01repeating_command_block\0\x01glass\0\x01iron_block\0\x1E\x06\x1D\x05\x07\0\0\0\0\x10\x1B\0\x01\0\0\x01kill @e[type=tnt]\0Kill TNT!\0\0\0\0\0\0\x01\x01\0\0\x1D\x09\x19\0\x1D\x4B\x3C\x15\0\x01\xBF\x4F\x07\0\x02\0\0\x1E\xF6\x07\0\x03\0\0XE`

The pseudo assembly code form of this file is:

```assembly
author 'DEMO\0'
addToBlockPalette 'tnt\0' ; ID: 0
addSmallX 3 ; brushPosition: {3,0,0}
addToBlockPalette 'repeating_command_block\0' ; ID: 1
addToBlockPalette 'glass\0' ; ID: 2
addToBlockPalette 'iron_block\0' ; ID: 3
addSmallZ 6 ; brushPosition: {3,0,6}
addSmallY 5 ; brushPosition: {3,5,6}
placeBlock (int16_t)0, (int16_t)0 ; TNT Block will be put at {3,5,6}
NewYadd ; *Y++, brushPosition: {3,6,6}
placeCommandBlockWithData (int16_t)1, (int16_t)0, 1, 'kill @e[type=tnt]\0', 'Kill TNT!\0', '\0', (int32_t)0, 1, 1, 0, 0 ; A command block will be put at {3,6,6}
addSmallY 9 ; brushPosition: {3,15,6}
addBigZ 1919804 ; 1919810: 00 1D 4B 3C = 01d4b3ch, brushPosition: {3,15,1919810}
addBigX 114511 ; 114511: 00 01 BF 4F = 01bf4fh, brushPosition: {114514,15,1919810}
placeBlock (int16_t)2,(int16_t)0 ; A glass block will be put at {114514,15,1919810}
addSmallZ -10 ; -10: F6 = 0f6h, brushPosition: {114514,15,1919800}
placeBlock (int16_t)3,(int16_t)0 ; A iron block will be put at {114514,15,1919800}
end
db 'E'
```

## Signing

*PhoenixBuilder* 0.3.5 implemented a bdump file signing system in order to identify the file's **real** publisher. Though using the PGP to sign is a good and secure way, we've chosen a signing method that highly depends on our authentication server since it's meaningless to implement the PGP signing just for an online program that can connect to the server anytime.

Note that a signature isn't required for a `bdx` file unless the user sets a `-S`(strict) flag. If you implemented the signing process, you should make sure that it works correctly since a `bdx` file with an incorrect signature won't be able to be processed by *PhoenixBuilder*.

### API

First let's learn the APIs of `bdx` file signing. We've implemented two apis to finish the signing process.

The host of those APIs is `uc.fastbuilder.pro` and HTTPS is required.

#### Signing

* Request:

    ```http
    POST /signbdx.web HTTP/1.1
    Host: uc.fastbuilder.pro
    User-Agent: MyApplication/0.1
    
    {"hash": "<The hash of your uncompressed bdx content without the end command 'X'.>","token": "<Your FastBuilder Token>"}
    ```

* Response:

  ```http
  HTTP/1.1 200 OK
  Content-Type: application/json
  
  {"success":true,"sign":"<Base64 of signature>",message:""}
  ```

#### Verifying

* Request:

    ```http
    POST /verifybdx.web HTTP/1.1
    Host: uc.fastbuilder.pro
    User-Agent: MyApplication/0.1
    
    {"hash": "<The hash of your uncompressed bdx content without the end command 'X'.>","sign": "<The signature's base64 value>"}
    ```

* Response:

  ```http
  HTTP/1.1 200 OK
  Content-Type: application/json
  
  {"success":true,"corrupted":false,"username":"<The person who signed the file>",message:""}
  ```

Note: After requesting the signing api, the base64 value of the signature should be decoded and written to the compressed part of the file, with the signature length and `isSigned` flag followed.
