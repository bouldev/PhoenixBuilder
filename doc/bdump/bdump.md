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

| ID                | Internal name                                             | Description                                                  | Arguments                                                    |
| ----------------- | --------------------------------------------------------- | ------------------------------------------------------------ | ------------------------------------------------------------ |
| 1                 | `CreateConstantString`                                    | Add the specified string to the palette, and the ID of the string is sorted in the term of the command called, e.g. the ID when assigning the first string given is `0`, and the ID of the second time is `1`. The maximum count of strings is `65536`. | `char *constantString`                                       |
| 2                 | **DEPRECATED and REMOVED**                                | -                                                            | -                                                            |
| 3                 | **DEPRECATED and REMOVED**                                | -                                                            | -                                                            |
| 4                 | **DEPRECATED and REMOVED**                                | -                                                            | -                                                            |
| 5                 | **DEPRECATED and REMOVED**                                | -                                                            | -                                                            |
| 6                 | `AddInt16ZValue0`                                         | Add `value` to the brush position's `Z` value, does the same thing as `AddInt16ZValue`. | `unsigned short value`                                       |
| 7                 | `PlaceBlock`                                              | Place a block on the current position of the brush using the ID of the string indicating the block's name returned by  `CreateConstantString` command and the `BlockData` value indicating the data value for the block. | `unsigned short blockConstantStringID`<br/>`unsigned short blockData` |
| 8                 | `AddZValue0`                                              | Add `1` to the brush position's `Z` value, does the same thing as the command `AddZValue`. | -                                                            |
| 9                 | `NoOperation`                                             | Do nothing. (No Operation)                                   | -                                                            |
| 10, `0x0A`        | **DEPRECATED and REMOVED**                                | -                                                            | -                                                            |
| 11, `0x0B`        | **DEPRECATED and REMOVED**                                | -                                                            | -                                                            |
| 12, `0x0C`        | `AddInt32ZValue0`                                         | Add `value` to the brush position's `Z`, does the same thing as `AddInt32ZValue`. | `unsigned int value`                                         |
| 13, `0x0D`        | `PlaceBlockWithBlockStates`                               | Place a block on the current position of the brush using the ID of the string indicating the block's name returned by  `CreateConstantString` command and the `BlockStates` string indicating the block states field for placing the block with `setblock` command in Minecraft. <br/>An example of `blockStates` is `["color":"orange"]` | `unsigned short blockConstantStringID`<br/>`char *blockStatesString` |
| 14, `0x0E`        | `AddXValue`                                               | Add `1` to the brush position's `X` value.                   | -                                                            |
| 15, `0x0F`        | `SubtractXValue`                                          | Subtract `1` from the brush position's `X` value.            | -                                                            |
| 16, `0x10`        | `AddYValue`                                               | Add `1` to the brush position's `Y` value.                   | -                                                            |
| 17, `0x11`        | `SubtractYValue`                                          | Subtract `1` from the brush position's `Y` value.            | -                                                            |
| 18, `0x12`        | `AddZValue`                                               | Add `1` to the brush position's `Z` value.                   | -                                                            |
| 19, `0x13`        | `SubtractZValue`                                          | Subtract `1` from the brush position's `Z` value.            | -                                                            |
| 20, `0x14`        | `AddInt16XValue`                                          | Add `value` to the brush position's `X`. `x` could be either positive, negative or zero. | `short value`                                                |
| 21, `0x15`        | `AddInt32XValue`                                          | Add `value` to the brush position's `X`. The difference between this command and the previous one is this command uses `int32_t` as its argument. | `int value`                                                  |
| 22, `0x16`        | `AddInt16YValue`                                          | Add `value` to the brush position's `Y`.                     | `short value`                                                |
| 23, `0x17`        | `AddInt32YValue`                                          | Add `value` to the brush position's `Y`.                     | `int value`                                                  |
| 24, `0x18`        | `AddInt16ZValue`                                          | Add `value` to the brush position's `Z`.                     | `short value`                                                |
| 25, `0x19`        | `AddInt32ZValue`                                          | Add `value` to the brush position's `Z`.                     | `int value`                                                  |
| 26, `0x1A`        | `SetCommandBlockData`                                     | Set the command block data for the block at the brush's position. **(Recommended to use command 36 instead)** | `unsigned int mode {Impulse=0, Repeat=1, Chain=2}`<br/>`char *command`<br/>`char *customName`<br/>`char *lastOutput (no effect and can be set to'\0')`<br/>`int tickdelay`<br/>`bool executeOnFirstTick`<br/>`bool trackOutput`<br/>`bool conditional`<br/>`bool needsRedstone` |
| 27, `0x1B`        | `PlaceBlockWithCommandBlockData`                          | Place a command block, and set its data at the brush's position. **(Recommended to use command 36 instead)** | `unsigned short blockConstantStringID`<br/>`unsigned short blockData`<br/>`unsigned int mode {Impulse=0, Repeat=1, Chain=2}`<br/>`char *command`<br/>`char *customName`<br/>`char *lastOutput (no effect and can be set to'\0')`<br/>`int tickdelay`<br/>`bool executeOnFirstTick`<br/>`bool trackOutput`<br/>`bool conditional`<br/>`bool needRedstone` |
| 28, `0x1C`        | `AddInt8XValue`                                           | Add `value` to the brush position's `X`. The difference between this command and the `AddInt16XValue` command is that this command uses `char` as its argument. | `char value //int8_t value`                                  |
| 29, `0x1D`        | `AddInt8YValue`                                           | Add `value` to the brush position's `Y`.                     | `char value //int8_t value`                                  |
| 30, `0x1E`        | `AddInt8ZValue`                                           | Add `value` to the brush position's `Z`.                     | `char value //int8_t value`                                  |
| 31, `0x1F`        | `UseRuntimeIDPool`                                        | Use a preset runtime id palette. `presetId` is the id of the runtime id palette used, which is assigned by PhoenixBuilder itself. The `presetId` for the current version of NetEase's Minecraft BE (1.17.0 @ 2.0.5) is `117`. Each runtime id matches a individual block state (contains its data value)<br/>See [fastbuilder/world_provider/runtimeIds.json](fastbuilder/world_provider/runtimeIds.json) for detailed content. **No longer being used** | `unsigned char poolId`                                       |
| 32, `0x20`        | `PlaceRuntimeBlock`                                       | Place a block with a specific runtime id at the brush's position. | `unsigned short runtimeId`                                   |
| 33, `0x21`        | `PlaceRuntimeBlockWithUint32RuntimeID`                    | Place a block with a specific runtime id at the brush's position. | `unsigned int runtimeId`                                     |
| 34, `0x22`        | `PlaceRuntimeBlockWithCommandBlockData`                   | Place a command block with the specified runtime id, and set its data at the brush's position. | `unsigned short runtimeId`<br/>`unsigned int mode {Impulse=0, Repeat=1, Chain=2}`<br/>`char *command`<br/>`char *customName`<br/>`char *lastOutput (no effect and can be set to'\0')`<br/>`int tickdelay`<br/>`bool executeOnFirstTick`<br/>`bool trackOutput`<br/>`bool conditional`<br/>`bool needRedstone` |
| 35, `0x23`        | `PlaceRuntimeBlockWithCommandBlockDataAndUint32RuntimeID` | Place a command block with the specified runtime id, and set its data at the brush's position. | `unsigned int runtimeId`<br/>`unsigned int mode {Impulse=0, Repeat=1, Chain=2}`<br/>`char *command`<br/>`char *customName`<br/>`char *lastOutput (no effect and can be set to'\0')`<br/>`int tickdelay`<br/>`bool executeOnFirstTick`<br/>`bool trackOutput`<br/>`bool conditional`<br/>`bool needRedstone` |
| 36, `0x24`        | `PlaceCommandBlockWithCommandBlockData`                   | Place a command block with the specified data value, and set its data at the brush's position. | `unsigned short data`<br/>`unsigned int mode {Impulse=0, Repeat=1, Chain=2}`<br/>`char *command`<br/>`char *customName`<br/>`char *lastOutput (no effect and can be set to'\0')`<br/>`int tickdelay`<br/>`bool executeOnFirstTick`<br/>`bool trackOutput`<br/>`bool conditional`<br/>`bool needRedstone` |
| 37, `0x25`        | `PlaceRuntimeBlockWithChestData`                          | Place a (chest) block with the specified data within the chest. `slotCount`'s type is `unsigned char` since Minecraft uses only a byte for the slot ID. | `unsigned short runtimeId`<br/>`unsigned char slotCount`<br/>`struct ChestData data` |
| 38, `0x26`        | `PlaceRuntimeBlockWithChestDataAndUint32RuntimeID`        | Place a (chest) block with the specified data within the chest. `slotCount`'s type is `unsigned char` since Minecraft uses only a byte for the slot ID. | `unsigned int runtimeId`<br/>`unsigned char slotCount`<br/>`struct ChestData data` |
| 39, `0x27`        | `RecordBlockEntityData`                    | Record the `Block Entity Data` where the brush is, and of course that other information can also be recorded.<br/>`uint32_t length` is refer to the length of `unsigned char buffer[length]` , and `unsigned char buffer[length]` itself is used to record some infomation.<br/>Due to some limitations, `PhoenixBuilder` can not record complete `NBT` data | `uint32_t length`<br/>`unsigned char buffer[length]` |
| 88, `'X'`, `0x58` | `Terminate`                                               | Stop reading. Note that though the general end is "XE" (2 bytes long), but a 'X' (1 byte long) character is enough. | -                                                            |
| 90, `0x5A`        | `isSigned` (fake command)                                 | A command that functions a little different with other commands, its argument is the previous byte of it, would only appear in the end of the file. Please do not use it unless you know how to use since an invalid signature would prevent PhoenixBuilder from constructing your structure. See paragraph `Signing` for details. | `unsigned char signatureSize`                                |

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
CreateConstantString 'tnt\0' ; ID: 0
AddInt8XValue 3 ; brushPosition: {3,0,0}
CreateConstantString 'repeating_command_block\0' ; ID: 1
CreateConstantString 'glass\0' ; ID: 2
CreateConstantString 'iron_block\0' ; ID: 3
AddInt8ZValue 6 ; brushPosition: {3,0,6}
AddInt8YValue 5 ; brushPosition: {3,5,6}
PlaceBlock (int16_t)0, (int16_t)0 ; TNT Block will be put at {3,5,6}
AddYValue ; brushPosition: {3,6,6}
PlaceCommandBlockWithCommandBlockData (int16_t)1, (int16_t)0, 1, 'kill @e[type=tnt]\0', 'Kill TNT!\0', '\0', (int32_t)0, 1, 1, 0, 0 ; A command block will be put at {3,6,6}
AddInt8YValue 9 ; brushPosition: {3,15,6}
AddInt32ZValue 1919804 ; 1919810: 00 1D 4B 3C = 01d4b3ch, brushPosition: {3,15,1919810}
AddInt32XValue 114511 ; 114511: 00 01 BF 4F = 01bf4fh, brushPosition: {114514,15,1919810}
PlaceBlock (int16_t)2,(int16_t)0 ; A glass block will be put at {114514,15,1919810}
AddInt8ZValue -10 ; -10: F6 = 0f6h, brushPosition: {114514,15,1919800}
PlaceBlock (int16_t)3,(int16_t)0 ; A iron block will be put at {114514,15,1919800}
Terminate
db 'E'
```

## Signing

*PhoenixBuilder* 0.3.5 implemented a bdump file signing system in order to identify the file's **real** publisher.

Note that a signature isn't required for a `bdx` file unless the user sets a `-S`(strict) flag. If you implemented the signing process, you should make sure that it works correctly since a `bdx` file with an incorrect signature won't be able to be processed by *PhoenixBuilder*.

We use hash method based on RSA for file signing. The server will issue an individual certification set for each user, and *PhoenixBuilder* signs the file with the private key and provide the public key chained the root key in the file, whose reality will be checked with the hardcoded server public key.

See `fastbuilder/bdump/utils.go` : `SignBDXNew`/`VerifyBDXNew` to see details.