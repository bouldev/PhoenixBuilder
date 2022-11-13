# Global Functions
## `printf([format][, ...args])`
* `format` `<any>`
* `...args` `<any>`

Prints formatted message to `stdout` **without a newline**.
```javascript
const count = 5;
printf("count: %d\n", count);
// Prints: count: 5, to stdout
```
<!-- Partially copied from the documentation of Node.JS -->
## `sprintf([format][, ...args])`
* `format` `<any>`
* `...args` `<any>`
* Returns: `<string>` The formatted string.
```javascript
let str=sprintf("val: %s","2");
// str = "val: 2"
```

## `require(name)`
Alias of [module.require](module.md#modulerequirename).
