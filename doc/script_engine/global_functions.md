# Global Functions

## `printf([format][, ...args])`
* `format` [<any>](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Data_structures#Data_types)
* `...args` [<any>](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Data_structures#Data_types)
Prints formatted message to `stdout` **without a newline**.
```
const count = 5;
printf("count: %d\n", count);
// Prints: count: 5, to stdout
```
<!-- Partially copied from the documentation of Node.JS -->
## `sprintf([format][, ...args])`
* `format` [<any>](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Data_structures#Data_types)
* `...args` [<any>](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Data_structures#Data_types)
* Returns: [<string>](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Data_structures#string_type) The formatted string.
```
let str=sprintf("val: %s","2");
// str = "val: 2"
```

## `require(name)`
Alias of [module.require](module.md#modulerequirename).
