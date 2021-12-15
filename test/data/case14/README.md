# Case 14

The destination folder is not owned by the process running the `Copy` operation, but is writeable by that process.

In this case, the Copy operation requires the `Options.PermissionControl = DoNothing` option to be set, otherwise a `chmod` operation is attempted on the destination dir, which will fail.

See https://github.com/otiai10/copy/pull/69.

`PermissionControl` can replace `AddPermission` option and `AddPermission` option is obsolete.
Use following instead:

```go
opt.PermissionControl = AddPermission(0222)
```