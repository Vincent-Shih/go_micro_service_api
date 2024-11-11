# auth Service

## Introduction

`auth Service` 是負責驗證使用權限的服務。正式環境下應只對內部群集開放。使用gRPC作為服務間的通訊協定

---

## 目錄結構

```tree
.
├── README.md
├── application
│   └── oauth_service.go
├── config
│   └── config.go
├── doc
├── domain
│   ├── aggregate
│   ├── entity
│   ├── repository
│   └── service
├── infrastructure
│   ├── grpcserver
│   │   └── server.go
│   └── repo_impl
├── main.go
└── migration
```

說明:

- root: 會有一個`main.go`當作服務的啟動點, `.env` 或 `docker-compose.yaml`...與專案無直接相關的檔案先放在跟資料夾下。是否要將build相關的檔案另開資料夾，可以討論
- /application: domain層對外溝通的實現放在這裡。如具體實現gprc service接口、dto。
- /config: 放置config檔相關文件
- /doc: openapi或相關文件都放這裡
- /domain: 業務邏輯的核心所在
- /infrastructure: 基礎設施
- /migration: db相關版本計畫放這裡

---