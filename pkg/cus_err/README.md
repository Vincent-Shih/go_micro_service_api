# Kgs Error

## Usage

Create a KgsErr

``` go
func foo() *cus_err.CusError {
    // Case 1.
   kgsErr := cus_err.New(cus_err.InternalServerError, "Your error message")

   // Case 2: New a cus_err with other error
   otherErr := someService()
   kgsErr = cus_err.New(cus_err.InternalServerError, "Your error message",otherErr)

   return kgsErr
}
```

Compare with kgsCode

``` go
func foo() {
    kgsErr := cus_err.New(cus_err.InternalServerError, "Your error message")
    if kgsErr.Code() == cus_err.InternalServerError {
        // Do something...
    }
}
```

Log a kgsError 

``` go
func foo(ctx context.Context) {
    kgsErr := cus_err.New(cus_err.InternalServerError, "Your error message")

    // If you want some field to log also, use `NewField()` to record with your err
    cus_otel.Error(ctx, err.Message(), cus_otel.NewField("token", req.AccessToken))

    // Or you can just send log simply
    cus_otel.Error(ctx, err.Error())
}
```
