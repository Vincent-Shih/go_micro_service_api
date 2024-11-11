# Go Micro Service Api

## 專案內容

1. frontend_api 客戶端api (http)
2. auth_service 安全檢查微服務 (gRpc)
3. user_service 用戶微服務 (gRpc)
4. 客製化lib
   1. rate limiter demo
   2. error code demo
   3. gin middleware demo
   4. database interface demo
   5. ent orm demo

## 操作說明

1. 使用 `postgresql.yml` 啟動 postgresql，並且使用帳號,密碼 `admin` 登入後創建 `auth`, `user` 資料庫

    ```sh
        docker compose -f postgresql.yml up -d
    ```

2. 使用 `redis.yml` 啟動 redis，密碼為 `P@ssw0rd`

    ```sh
        ```sh
        docker compose -f redis.yml up -d
    ```

3. 使用 `grafana.yml` 啟動 grafana, loki, tempo, prometheus, open telemetry

    ```sh
        docker compose -f grafana.yml up -d
    ```

## 服務流程

1. 由 `frontend_api` 作為給前端的進入點
2. `auth_service` 作為安全認證，以及在 frontend_api 的middleware 做安全認證時會請求的部分
3. `user_service` 基於用戶相關資料的服務
4. 以上服務希望做到盡可能的放每個服務只做自己的事情，frontend_api 進行業務流程的控制