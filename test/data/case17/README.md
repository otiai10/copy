So if you wanted to ignore error you should add something like this:
```go
opt.OnFileErr = func(_ error) error { return nil } 
```
The default value is nil and accepts raised error.