# AWS S3 client

AWS S3 的抽象層

<!-- @import "[TOC]" {cmd="toc" depthFrom=1 depthTo=6 orderedList=false} -->

<!-- code_chunk_output -->

- [AWS S3 client](#aws-s3-client)
  - [Usage](#usage)
    - [ListObjects](#listobjects)
    - [GetObject](#getobject)
    - [PutObject](#putobject)
    - [DeleteObject](#deleteobject)

<!-- /code_chunk_output -->

## Usage

!!! warning 預先準備
    使用前要先把 AWS 的授權檔案`~/.aws/config`, `~/.aws/credentials`準備好，或者寫在環境變數

以下為手動建立服務端

```go
ctx := context.Background()
cfg, err := s3.NewAWSConfig(ctx)
if err != nil {
    panic(err)
}
client := s3.NewClient(cfg)
service := s3.NewService(client)
```

或者用 FX

```go
fx.Provide(
    s3.NewS3ClientFx(),
)
```

---------------

目前開放四種功能

### ListObjects

可以列出bucket內的物件，再用`prefix`跟`startAfter`做篩選

```go
listResult, err := service.ListObjects(ctx, &bucket, &prefix, nil)
if err != nil {
    panic(err)
}
for _, obj := range listResult.Contents {
    println(aws.ToString(obj.Key), obj.Size)
}
```

### GetObject

利用bucket跟key，取得個別物件

```go
getResult, err := service.GetObject(ctx, &bucket, &key)
if err != nil {
    panic(err)
}
defer getResult.Body.Close()
println(io.ReadAll(getResult.Body))
```

### PutObject

更新bucket內key為key的物件

```go
putResult, err := service.PutObject(ctx, &bucket, &key, strings.NewReader("Hello, World!"))
if err != nil {
    panic(err)
}
println(putResult.VersionId)
```

### DeleteObject

刪除bucket內key為key的物件

```go
err = service.DeleteObject(ctx, &bucket, &key)
if err != nil {
    panic(err)
}
```

範例在 `examples/main.go`

不過沒辦法實際使用，所以不確定是否正確😅
