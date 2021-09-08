# 自用的 log 库

简单使用:

```go
option := Option{
    WriteToStd:  false,
    WriteToFile: true,
    OutputLevel: "debug",
}
Init(option)
```

