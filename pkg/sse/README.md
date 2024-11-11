# Sever-send-event middleware

SSE的中間層

## Usage

### Publisher

首先建立一個stream物件，會註冊該應用程式的串流客戶端

```go
s := sse.NewStream()
```

接著建立一個端點，SSE功能如下，需要header跟stream的middleware
前者告訴瀏覽器是串流，後者註冊到stream物件

訊息通道可由 `sse.GetChannel(c)` 從context中取出
使用gin內建的`Stream`啟動串流，如果從訊息通道讀到訊息，就用`SSEvent`生成response

```go
r := gin.Default()

// publish endpoint
r.GET("/sse", sse.HeadersMiddleware(), sse.StreamMiddleware(s), func(c *gin.Context) {
    clientChan := sse.GetChannel(c)
    c.Stream(func(w io.Writer) bool {
        // Stream message to client from message channel
        if msg, ok := <-clientChan; ok {
            c.SSEvent("message", msg)
            return true
        }
        return false
    })
})
```

這邊有個假的訊息提供者，每秒打出訊息到stream物件的訊息通道
可視為廣播，目前沒有額外路由管理

```go
// publisher
go func() {
    for {
        time.Sleep(time.Second * 1)
        now := time.Now().Format("2006-01-02 15:04:05")
        currentTime := fmt.Sprintf("The Current Time Is %v", now)

        // Send current time to clients message channel
        s.Messages <- currentTime
    }
}()
```

### Subscriber

建立client

有提供瀏覽器以及純golang的範例

先發出GET請求

!!! warning 記得關閉 response body

回應內容會是串流，所以需要scanner接收

```go
// plain get is able to stream the data
res, err := http.Get("http://localhost:8080/sse")
if err != nil {
    panic(err)
}
defer res.Body.Close()

scanner := bufio.NewScanner(res.Body)

// keep streaming out
for scanner.Scan() {
    println(scanner.Text())
}

if err := scanner.Err(); err != nil {
    panic(err)
}
```

瀏覽器可用`EventSource`或`fetch streaming`

可以單獨點擊`index.html`或直接打開 [http://localhost:8080](http://localhost:8080)

!!! note
    瀏覽器有限制每個來源最多只有6個SSE連線(http1)，100個連線(http2)

```js
const source = new EventSource('http://localhost:8080/sse');
source.onopen = function () {
    console.log('Connection was opened.');
};
source.onmessage = function (event) {
    const node = document.createElement('p');
    node.textContent = event.data;
    document.body.appendChild(node);
};
source.onerror = function (error) {
    console.error('Error occurred:', error);
};
```
