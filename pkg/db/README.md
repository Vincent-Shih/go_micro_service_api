# Database

作為 Database 的抽象層

## Usage

### Application層

如果是有更動操作(如：Create, Update, Delete, .etc)，需要使用如下套版

一律begin，遇到錯誤就手動rollback，全部完成就commit

```go
// begin transaction
ctx, err := s.db.Begin(ctx)
if err != nil {
    return nil, err
}

// 更動操作(如：Create, Update, Delete, .etc)
err = some_modified_database_operation.....

// if error, rollback transaction
if err != nil {
    _, rollbackErr := s.db.Rollback(ctx)
    if rollbackErr != nil {
        cus_otel.Error(ctx, rollbackErr.Error())
        err = rollbackErr
    }
    return nil, err
}

// Commit the transaction
_, commitErr := s.db.Commit(ctx)
if commitErr != nil {
    cus_otel.Error(ctx, commitErr.Error())
    return nil, commitErr
}

```

### Repository層

如果是有更動操作(如：Create, Update, Delete, .etc)，需要使用如下套版

一律使用tx進行操作

```go
tx, ok := repo.db.GetTx(ctx).(*ent.Tx)
if !ok {
    return nil, cus_err.New(cus_err.InternalServerError, "failed to get transaction", nil)
}

// TODO: tx.do_database_operation
```

如果是有 讀取 操作，需要使用如下套版

一律先`GetTx`，找不到就用`GetConn`進行操作

可使用`GetClient`的封裝

```go
client := repo.db.GetClient(ctx).(*ent.Client)

// TODO: client.do_database_operation
```
