# CloudChest

A cloud storage web application made out of curiosity and also a hobby project. This web app allows users to store, view, and modify their personal files on the cloud.

Tech stack used:
- Vue.js (Front-end)
- Gin (Back-end)
- MySQL (RDBMS)
- MinIO (Object Storage)

## Known Issues

### Grandparent folder is not restored
When a file within a folder hierarchy (e.g., file <- parent folder <- grandparent folder) is deleted, followed by deleting its parent and grandparent folders, attempting to restore the file from the trash will not restore the grandparent folder. This happens because recursive checks to restore parent folders only trigger if the immediate parent folder is also deleted.

```
file <- parent folder <- grandparent folder
```
IF we delete `file`, and later delete `grandparent folder`
```
file (deleted) <- parent folder <- grandparent folder (deleted)
```
Attempting to restore `file (deleted)` **WILL NOT** restore `grandparent folder (deleted)`, thus ending up with this state
```
file <- parent folder <- grandparent folder (deleted)
```
This will make the `parent folder` and `file` not showing up in the file explorer. The current workaround is to also restore `grandparent folder (deleted)`

> Fixes are planned in the future.

## Deployment Guide