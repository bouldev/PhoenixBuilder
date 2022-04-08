# Game

## `game.eval(command)`
* `command` [<string>](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Data_structures#string_type) PhoenixBuilder's command
`game.eval()` executes a PhoenixBuilder's command.
```
game.eval("get");
game.eval("round -r 10");
```

## `game.oneShotCommand(command)`
* `command` [<string>](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Data_structures#string_type)
`game.oneShotCommand()` executes a Minecraft command without waiting for response.
```
game.oneShotCommand("kill @a");
```

## `game.sendCommandSync(command)`
* `command` [<string>](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Data_structures#string_type)
* Returns: [<Object>](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Object)
Executes `command` in game and wait until receiving the response.
** Warning: For commands without a response, this command will lead your script into a deadlock. **

## `game.sendCommand(command[, callback])`
* `command` [<string>](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Data_structures#string_type)
* `callback` [<Function>](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Function)
** `response` [<Object>](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Object)
`game.sendCommand()` executes `command` and call `callback` once the response is received.
Same as [game.oneShotCommand](game.md#gameoneshotcommandcommand) when `callback` is not assigned.

## `game.botPos()`
* Returns: `ret` [<Object>](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Object)
** `x` `<Number>`
** `y` `<Number>`
** `z` `<Number>`
Get the current position of the bot.

## `game.subscribePacket(packetType, callback)`
* `packetType` `<string>` One of the packet type in fastbuilder/script_engine/packetType.go
* `callback` `<Function>` The callback that will be called once the packet with the specified type is received.
** `packet` `<Object>`
* Returns: `<Function>` The function to unsubscribe the packet

## `game.listenChat(callback)`
* `callback` `<Function>`
** `name` `<string>`
** `message` `<string>`

