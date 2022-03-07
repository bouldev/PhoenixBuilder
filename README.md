# FastBuilder Phoenix 

## Description

> FastBuilder is a structure generating tool for Minecraft Bedrock Edition that supporting various platforms. The Phoenix ver. was designed for the rental server of the Minecraft Netease Edition. Currently supports generating Euclid geometry and ACME(mcacblock) · schematic file(NBT data of blocks would be disposed)'s structures, and images painting.

**NOTE: FastBuilder Phoenix is a commercial software!**

### Principle

FastBuilder currently using an all-new technology, and no longer limited by the WebSocket. Thanks to it, the all-new FastBuilder's speed, performance and stability got a significant improvement, and it became highly extendable. The core of the current version of FastBuilder is based on Sandertv's [Gophertunnel](https://github.com/Sandertv/gophertunnel/) project that licensed under the MIT License.

### Open Source

FastBuilder Phoenix's **client** is fully opened source [on GitHub](https://github.com/LNSSPsd/PhoenixBuilder).

(The source code is licensed under GPL v3.)

### Acknowledgements (Read before purchasing)

* Purchasing it represents you agree and would comply with the [FastBuilder User License](LICENSE.html).

- The technology that it uses may expire any time, so FastBuilder life-time support is not guaranteed.
- The version of FastBuilder you bought contains the features explained above, we may add other features in the future but we can't guarantee to that.
- Some steps of the installation and the use of FastBuilder needs a certain degree of **IT and Math knowledges**, and the instruction for installation would consider that you **have** these required knowledges by default.
- FastBuilder personal account can only be binded to at most two rental servers, and it allows to be changed once per month. FastBuilder Phoenix can't be used in multi-player mode of NEMC or in the international version of Minecraft.
- Please **do not** import any files **without the author's permission**, every single content creator survives in the community with their own energy and wisdom. Using others' IP (stands for Intellectual Property) commercially would cause destructions to the whole environment of the game. Moreover, if the copyright owner of contents you have used ascertains the liability, you should and must take your responsibility, and we shouldn't bear any liability for it.
- Developers aren't customer service reps, they don't have the obligation to **resolve** the various problems you met when using the software, moreover they would not guide you to install the software **by themselves**. If you found bugs when using the software and **you are very sure of it**, you can **submit bug reports** to a seller.
- The current version (Phoenix Alpha) is under testing, various bugs might be triggered, purchasing means you **volunteer to bear the risks**.
- **The nearly full-English prompts in the program doesn't mean that we are not Chinese.**
- We designed every feature **as perfect as possible**, but it still contains **defects**.

### World View

The differences between FastBuilder and other programs is that it contains the concept of "server" and "client". The game that players running is a client, FastBuilder is also a client, and the rental server the player enters is a server.

Therefore, clients can run on different devices, and FastBuilder can be executed to join your server w/o your game running.

The abilities explained below are required since FastBuilder needs the operations of command lines.

* The ability of operating files: Able to understand the level of paths and files.

- The ability of reading English: Able to identity words like "Error", "Permission denied" or "Not found" and know what they stands for.
- The ability to differ the full-width and the half-width characters.
- It's better to have the ability to enter and execute commands in the command line interface.

Please ensure that you have the abilities mentioned above, if you encountered problems when using FastBuilder because you are not satisfying the mentioned conditions, development group would not bear any liabilities, moreover it will not provide any help.

### Installation Instruction

#### Requirements

- Holds a FastBuilder User Center account, aka a FastBuilder account (bought FastBuilder).
- Depends on your device:
  * PC (Windows/Linux/macOS): Your computer should have a fully functional network adapter installed.
  * Android: Have *Termux* installed, see below for instruction.
  * iOS: Jailbroken, and you know how to use the terminal.
- Hard-working hands and thoughtful brain

1. Login to [FastBuilder User Center](https://uc.fastbuilder.pro) 
2. Click the *Profile* tab, and set the *<ruby><rb>Minecraft Netease Edition Username</rb><rp>(</rp><rt style="font-size:80%;">网易版用户名</rt><rp>)</rp></ruby>*.
3. Enter the number of the <ruby><rb>rental server</rb><rp>(</rp><rt style="font-size:80%;">租赁服</rt><rp>)</rp></ruby> that you want to use FastBuilder on.(note: The rental server should accept the entrance of **any <ruby><rb>level</rb><rp>(</rp><rt style="font-size:80%;">等级</rt><rp>)</rp></ruby>**'s player, satisfy it by turning off "<ruby><rb>player entrance level requirement</rb><rp>(</rp><rt style="font-size:80%;">玩家等级准入要求</rt><rp>)</rp></ruby>" toggle in the server settings interface)
4. Set a nickname of the <ruby><rb>helper user</rb><rp>(</rp><rt style="font-size:80%;">辅助用户</rt><rp>)</rp></ruby>, then click **[<ruby><rb>Create</rb><rp>(</rp><rt style="font-size:80%;">生成</rt><rp>)</rp></ruby>]** to create one.

That's all for the completion of necessary informations, and the following content is the steps of installation, different platforms have different solutions, please find your own platform:

#### Steps for Installation

- Windows: Download directly in the [Download] tab of the user center.

- iOS: Install the package from our APT package source.

- Linux x86_64 (recommended platform): 

  ```shell
  wget -O fastbuilder https://storage.fastbuilder.pro/epsilon/phoenixbuilder
  chmod +x fastbuilder
  ```

- Android: 

  - a. Click [here](https://f-droid.org/repo/com.termux_117.apk) to install Termux (**0.117**); (Or download it from a trustable source by yourself.)  

  - b. After the installation of Termux, navigate to your system configuration, and give it **the permission of accessing the storage space (aka sdcard)**, and allow it to **run in the background without limitations**.

   - c. Download the FastBuilder binary. (x86 or x86_64 android devices are not supported.)

     > **Note: This step (c) is also the way of upgrading FastBuilder Phoenix, execute this step directly to upgrade FastBuilder afterwards.**

     ```shell
     o=$(uname -o) a=$(uname -m) && if [ "$o" == "Android" ]; then [[ "$a" == "aarch64" ]] && f="arm64" || f="armv7" && curl -o fastbuilder https://storage.fastbuilder.pro/phoenixbuilder-android-executable-$f && chmod +x fastbuilder && ./fastbuilder; else echo "for Android only"; fi
     
     ```
     **Thanks [@CMA2401PT](https://github.com/CMA2401PT) for providing the easier version of the download command for fastbuilder.**

### Usage

FastBuilder Phoenix is a pure command line program ~~without complicated GUI~~, which made the program very easy to use.

#### Launching

- Windows: Double-click FastBuilder executable(**.exe**) file to execute.

- Linux: No need to explain.

- iOS: Open the terminal and execute the following command:

  ```shell
  fastbuilder
  ```

- Android: Open the Termux app and execute the following command:

  ```shell
  ./fastbuilder
  ```

#### Initialization

If no exceptions happened, after finishing these steps, you will see the FastBuilder's copyright notice and other things. It will ask you to enter your <ruby><rb>FBUC</rb><rp>(</rp><rt style="font-size:80%;">FastBuilder User Center</rt><rp>)</rp></ruby> username and password (**Password won't be echoed**), and you won't need to do that it twice.

Then, FastBuilder will ask you to enter the rental server number and its password(Press *Enter* directly if none, **won't be echoed**). If it haven't crashed, presumably it has been launched.

After that, leave it in the background, and enter the rental server. Seeing the helper user is online(in the user list in `/list` command or in pause interface) means that FastBuilder works properly. Please **give the helper user <ruby><rb>OP</rb><rp>(</rp><rt style="font-size:80%;">operator</rt><rp>)</rp></ruby> permission**. ~~The helper user will only listen to **<ruby><rb>operator's</rb><rp>(</rp><rt style="font-size:80%;">your</rt><rp>)</rp></ruby>** commands, so the the *Minecraft Netease Edition Username* should be set to the same to **you nickname in *Minecraft Netease Edition***. Please do not use skin packs with the **<ruby><rb>title</rb><rp>(</rp><rt style="font-size:80%;">称号</rt><rp>)</rp></ruby>** since the helper user won't be able to process your commands.~~ Please enter commands in the console since netease will ban accounts that entered fastbuilder commands in the chat scene. For that reason, it's also unrecommended to use the `get` command of FastBuilder as it gives the name of the controller to the backend, which may cause an auto ban.

#### FastBuilder Command Resolving

FastBuilder uses a system similar to Linux Shell (it isn't the same command system of Minecraft). The "/" is needless and you can execute it by simply send it as a chat message.

Note that you can't use "#" to give a comment in FastBuilder's commands.

```shell
# Example: Generate a round with radius 5, faces the y-axis.
# These 2 commands below do the same thing.
round -r 5 -f y -h 1
round --radius 5 --facing y --height 1
```

##### Generator Settings

After initializing FastBuilder, we call the dimension that the <ruby>helper user<rp>(</rp><rt style="font-size:80%;">bot</rt><rp>)</rp></ruby> in as a *space*, every operations would be executed in this space. So if you want to use FastBuilder in a different space, you should teleport the bot to the target space.

To use FastBuilder, you should set the **origin** of the **space**(**structures will be built around the origin**), and the default origin is the position where bot entered the game.

Use `get` command to modify the origin to the current position of **you**.

The usage of commands of corresponding features are shown below.

FastBuilder Phoenix used the **multi-task system**, which means that you can have more than one tasks to be executed at the same time, and you can use the `task` command to manage tasks.

##### Task Command

`task` command was designed for managing the current **tasks**(**building process**). Every task has its own runtime, and you can use some functional commands to set the internal arguments for tasks. `task` command has the following basic child commands(Texts after `#` are comments, which are used to explain what the child command for):

````shell
tasktype <type:async|sync> # **Global command**, set the task type for newly-created tasks. Sync mode doesn't support the progress displaying but builds at the same time of calculating, and async mode supports displaying the progress since it builds after calculation.
task list # Lists the ID of each tasks, and its content · status.
task pause <taskID> # Pauses a specific task
task resume <taskID> # Resumes a specific task
task break <taskID> # Destroys a specific task (Unrecoverable, the task will be gone from the task list)
````

`task` command can also be used to set the delay of a task or its delay mode:

```shell
task setdelay <taskID> <delay> # Sets the delay for the specified task. The unit in continuous mode is microsecond and it's second in discrete mode.
task setdelaymode <taskID> <delayMode:continuous|discrete|none> # Sets the delay mode for the specified task, there are 3 modes available.
task setdelaythreshold <taskID> <threshold:int> # Sets the threshold for a specific task, only available in discrete mode.
```

- continuous: Send packets with a specific speed.

- discrete: Send packets without delay after wait the time of `delay`'s value (the max count of packets per delay won't be greater than the threshold).

- none: Send packets continuously without delay.

Each mode has its own advantages and disadvantages, please handle is as you think fit:

* Speed: continuous <= discrete < none (For some special configurations, continuous mode's speed may faster than discrete mode's speed)
* Stability: continuous > discrete > none
* Smoothness: continuous > none >= discrete

##### Functional Commands

- Set the origin for the space:

  ```shell
  get
  set x y z
  ```

- Set the global command execution delay solution:

  ```shell
  delay mode <delayMode:continuous|discrete|none> # Sets the default packet sending solution
  delay threshold <threshold:int> # Sets the default threshold, only available in discrete mode.
  delay set <Delay:int> # Sets the default packet sending delay. The unit in continuous mode is microsecond and it's second in discrete mode.
  ```

- Set whether to show the progress (show informations like the percentage of construction, total block count, and the instantaneous velocity. default:`true`)

  ```shell
  progress <value:bool>
  ```

* Logout from FastBuilder User Center

  ```shell
  logout
  ```

* Reselect the preferred language

  ```
  lang
  ```

* Open the FastBuilder controlling menu

  ```
  menu
  ```

  

##### Geometric Commands

FastBuilder has the ability of constructing simple geometry structures in the space. (like round, circle, sphere, line, ellipsoid, etc.)

- Round/Circle:
  ```shell
  round/circle -r <radius> -f <facing> -h <height> -b <tileName> -d <data>
  -r --radius The radius of the round or circle.
  -f --facing Facing, available values: x,y,z
  -b --block Block to be used to construct the structure
  -d --data The data (aka damage value) of the block to be used to construct the structure
  ```

- Sphere:
  ```shell
  sphere -r <radius> -s <shape>
  -s --shape hollow|solid
  ```

- Ellipse:

  ```shell
  ellipse -l <length> -w <width> -f <facing>
  -l --length Length
  -w --width Width
  ```

- Ellipsoid:

  ```shell
  ellipsoid -l <length> -w <width> -h <height>
  ```

##### Structure Construction Commands

Load and construct structures from `schematic`, <ruby><rb><code>bdx</code></rb><rp>(</rp><rt style="font-size:80%">bdump</rt><rp>)</rp></ruby> or `mcacblock`(structure file exported by the `ACME` building tool) files:

```shell
schem -p <filePath>
acme -p <filePath>
bdump -p <filePath>
-p --path The path of the file, use "" to assign a path with blankspace(s).
# optional flags: --excludecommands : Exclude commands in command blocks
# 				  --invalidatecommands: Invalidate commands in command blocks by adding characters to the start of the command content, for example, the invalid form of "say 123" is "|say 123".
#                -S --strict       : Break if the file is unsigned or failed to verify its signature.
```

##### Painting Slice Construction

- Load image to the space:

  ```shell
  plot -p <imageFilePath:string> -f <facing:x|y|z>
  ```

##### Experimental: Structure Export

**WARNING: This feature is unstable, unexpected things might happen during the use of this feature. Please check whether the exported file is valid since sometimes it might be exported to an invalid format.**

* Set the start point of export

  ```shell
  get
  set x y z
  ```

  or

  ```shell
  get begin
  ```

  - These are two different forms of a same command.

* Set the end point of export

  ```shell
  get end
  setend x y z
  ```

* Export the structure in assigned area to a file

  ```shell
  export -p <filePath>
  # optional flag: --excludecommands : Exclude commands in command blocks
  ```

* Import the exported file with the command `bdump` mentioned above.
