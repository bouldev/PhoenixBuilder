# Engine

## `engine.setName(name)`
** Warning: For a script package, this shouldn't be called as the name will be set automatically **
* `name` [<string>](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Data_structures#string_type) The display name for the script.
Display name will be attached as the prefix of the script's output to user.
```
engine.setName("My Script");
```
## `engine.waitConnectionSync()`
Wait until the connection to the game is established.
## `engine.waitConnection(callback)`
* `callback` [<Function>](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Function)
Asynchronously wait until the connection to the game is established.

## `engine.message(message)`
** Deprecated: Use [printf](global_functions.md#printfformat-args) or [console.log](console.md#consolelogdata-args) instead. **
* `message` [<string>](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Data_structures#string_type)
Display a message.

## `engine.crash(reason)`
* `reason` [<string>](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Data_structures#string_type)
Throw an exception and terminate the execution of the script.

