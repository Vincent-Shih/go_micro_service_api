# Frontend Api

## Introduction

`Frontend Api`是對前端(Web)公開的統一接口。

## Rule

1. 接口風格盡量符合RESTful所定義的規範。
2. 一般情況Response Body採用json格式回傳，並符合與前端約定之格式。
3. Token參數由Header傳入與傳出。

## 結構目錄

```sh
.
├── api
│   └── v1
├── config
├── doc
├── infrastructure
│   ├── client
│   └── httpserver
├── middleware
│   └── auth
├── route
└── test
```

資料夾說明:

- root: 會有一個`main.go`當作服務的啟動點, `.env` 或 `docker-compose.yaml`...與專案無直接相關的檔案先放在跟資料夾下。是否要將build相關的檔案另開資料夾，可以討論
- /api: 放置各版本的的api實作內容(handler / controller)
- /config: 放置設定相關的檔案
- /doc: 裡面放置OpenApi或其他相關資料的文檔
- /infrastructure: 放置基礎設施 如: gRPC的Client DI的啟動函數 Repo的實作...等
- /middleware: 放置中間件
- /route: 放置路由綁定的檔案
- /test: 放置測試,如果有private function 或小型的uni_test要想做測試 也可將測試檔建立在該資料夾底下(沒有硬性規定 但盡量還是放在/test內)

## Specifications

為了確保大家的api是可被驗證、擁有統一的格式，規範大家的handler寫法，一般情況下的handler 須滿足。(Streaming 與 SSE 或 Websocket 長連線的的回覆不在此限內)

1. 在處理邏輯前 開啟一個Trace
2. 回覆時使用responder封裝成統一格式

範例：

``` go

//  test struct  just for this example
type test struct {
    Msg   string            `json:"msg"`
    Array []string          `json:"array"`
    H     map[string]string `json:"map"`
}

func fooHandler(c *gin.Context) {
    // Start trace before you start
    _, span := cus_otel.StartTrace(c.Request.Context())
    defer span.End()

    // Do some logic...

    // Use responder for unified reply
    responder.Ok(test{
        Msg:   "Hello, World!",
        Array: []string{"a", "b", "c"},
        H: map[string]string{
            "key1": "value1",
            "key2": "value2",
        },
    }).WithContext(c)
}
```

## Swagger

### 環境設定

首先我們需要先在電腦上安裝，`swag` 套件，方便我們日後進行 swagger 自動生成的操作

1. 安裝swag

   ```sh
    $go install github.com/swaggo/swag/cmd/swag@latest
   ```

2. 檢查是否正確安裝

    ```sh
    $swag -v
    ```

3. 這邊會在畫面中看到 `swag` 的版本，如果看到下面內容就算安裝成功

    ```sh
    swag version v1.16.3
    ```

### 設定 Router

1. 首先，因為 `swagger` 也是由我們所寫的 server 所渲染的畫面，所以路由的不分也需要參照伺服器的設定
2. 根據我們現在的專案，找到 `route` 目錄，在這裡建立一個 `swagger_route.go`

    ```go
    package route

    import (
        "github.com/gin-gonic/gin"
        swaggerFiles "github.com/swaggo/files"
        ginSwagger "github.com/swaggo/gin-swagger"
    )

    func addSwaggerV1Routers(r *gin.Engine) {
        r.GET("/swagger/v1/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
    }

    ```

3. 接下來，我們回到 `route.go` 這個檔案，在 `RegisterRouters` 這個方法中加入 `addSwaggerV1Routers(r)` ，可能有其他方法可以加入

    ```go
    func RegisterRoutes(r *gin.Engine) {
        addSwaggerV1Routers(r)
        v1 := r.Group("/api/v1")
        addTestRoutes(v1)
        addAuthRoutes(v1)
    }
    ```

4. 接著執行

   ```sh
   swag init
   ```

5. 這時候會看到 `frontend_api`，底下增加了一個 `docs` 目錄，這邊因為 swagger 預設的目錄即為 `docs`

### 針對 api 產生 swagger doc

1. 接下來我們，以 `internal/api/v1/oauth_handler.go` 為範例產生 swagger
2. 在檔案中 `ClientOauth` 是我們的 api 進入點，我們希望由 swagger 自動產生，所以我們必須要在方法的上方加上一些描述
3. 如下

    ```go
    // ShowAccount godoc <-- 這個只是給我們看到
    // @Router       /api/v1/oauth [get]    <--- 這個很重要這個就是告訴 swagger 這隻api 打的是哪個路由，以及請求的方法是 POST 或 GET
    // @Summary      sign in client by jwt <-- 這個會在 swagger 畫面中的 api 後方出現
    // @Description  validate client
    // @Tags         auth
    // @version      1.0
    // @Accept       json
    // @Produce      json
    // @Param        id   path      int  true  "Account ID"
    // @Success      200  {object}  object{code=int,traceId=string,data=v1.User}
    // @Failure      400  {string}  httputil.HTTPError
    // @Failure      404  {string}  httputil.HTTPError
    // @Failure      500  {string}  httputil.HTTPError
    ```

4. 上述的這些描述，不一一說明，要注意的重點為以下幾個
   1. `@Router` 這個就是告訴 swagger 這隻api 要打的位置以及 method
   2. `@Summary` 可以讓用的人快速知道這隻 `api` 是做什麼的
   3. `@Description` 可以針對api做更仔細的描述，當然不一定只會給前端使用，也可以把裡面的邏輯寫給組員看
   4. `@Tag` 這個會直接幫 api 進行分類，例如 `oauth` 底下有 get, valid 等都會被歸類在這個下方
   5. `@Param` 這個會直接在畫面中產生欄位，讓使用的人可以直接進行操作
   6. `@Success` 成功範例
   7. `@Failure` 失敗範例，這個可以有多個，如上面的描述可以知道總共有幾個錯誤，個別代表的是什麼樣的錯誤內容

5. 完整範例

    ```go
    package v1

    import (
        "go_micro_service_api/pkg/responder"

        "github.com/gin-gonic/gin"
    )

    type User struct {
        Id    string `json:"id" example:"1"`
        Email string `json:"email" example:"vincent@cus.tw"`
    }

    // ShowAccount godoc
    // @Summary      sign in client by jwt
    // @Description  validate client
    // @Tags         auth
    // @version      1.0
    // @Accept       json
    // @Produce      json
    // @Param        id   path      int  true  "Account ID"
    // @Success      200  {object}  object{code=int,traceId=string,data=v1.User}
    // @Failure      400  {string}  httputil.HTTPError
    // @Failure      404  {string}  httputil.HTTPError
    // @Failure      500  {string}  httputil.HTTPError
    // @Router       /api/v1/oauth [get]
    func ClientOauth(c *gin.Context) {
        user := User{
            Id:    "1",
            Email: "9BQ9n@example.com",
        }
        responder.Ok(user).WithContext(c)
    }
    ```

6. 接下來執行執行以下指令

   ```sh
   swag init -g ./internal/api/v1/*.go
   ```

7. 如此就會直接把 `internal/api/v1` 之下的所有檔案掃描一遍，並且產生文件已便 swagger 做畫面宣染
8. 接下來我們就可以啟動服務，根據我們設定的路由 swagger 生成[網址]( http://localhost:8080/swagger/v1/index.html)

   ```url
    http://localhost:8080/swagger/v1/index.html
   ```

9. 操作畫面如下
   ![images](../asserts/swagger_sample.png)

10. 如果在api 上面對 swagger 進行描述值如果使用到專案外部的 `struct`，請記得在要外層執行，不要在 frontend_api中執行

    ```sh
    swag init -g frontend_api/internal/api/v1/oauth_handler.go  -o ./frontend_api/docs/
    ```

11. 另外，如果請求 swagger網址的時候，發生 `401`，請記得調整 `authMiddleware`

## 參考連結

1. [swaggo/swag](https://github.com/swaggo/swag)