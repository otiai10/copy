So if you wanted to limit copy rate with a KB/second value, you could add options like this:
```go
opt.CopyRateLimit = 100
```
The default value is 0, which means no limit.