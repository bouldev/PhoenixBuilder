# Engine
## `engine.setName(name)`
** Warning: For a script package, this shouldn't be called since the name will be set automatically **
* `name` `<string>` The display name for the script.

Display name will be attached as the prefix of the script's output to user.
```javascript
engine.setName("My Script");
```
## `engine.waitConnectionSync()`
Wait until the connection to the game is established.
## `engine.waitConnection(callback)`
* `callback` `<Function>`

Asynchronously wait until the connection to the game is established.

## `engine.message(message)`
** Deprecated: Use [printf](global_functions.md#printfformat-args) or [console.log](console.md#consolelogdata-args) instead. **
* `message` `<string>`

Display a message.

## `engine.crash(reason)`
* `reason` `<string>`

Throw an exception and terminate the execution of the script.

