# File System
## `fs.containerPath`
* Type: `<string>`

The absolute path of the container for the current script or package. Will be `""` if no container is created.
For a script package, the container will be created automatically unless `manifest.no_container` is set.

## `fs.requireContainer(container_identifier)`
* `container_identifier` `<string>` The identifier for the container. Only **english characters, numbers, and `.`, `_`, `-`** are allowed for the identifier. And the identifier's length should between 5 characters to 31 characters. (`32>len>4`)

Requires a container for the script. Duplicated call will cause an exception to be thrown.
**For a script package, this should NOT be called**.

## `fs.exists(path)`
* `path` `<string>` The path to the file, can be absolute or relative path (to the script container).
* Returns: `<Boolean>` A boolean value indicates the file's existence.

## `fs.isDir(path)`
* `path` `<string>`
* Returns: `<Boolean>` A boolean value indicates whether the specified item is a directory.

## `fs.mkdir(path)`
* `path` `<string>`

Make a directory. Intermediate directories will be created.

## `fs.rename(oldpath, newpath)`
* `oldpath` `<string>`
* `newpath` `<string>`

## `fs.remove(path)`
* `path` `<string>`

## `fs.readFile(path)`
* `path` `<string>
* Returns: `<string>` File content.

## `fs.writeFile(path, content)`
* `path` `<string>`
* `content` `<string>`

