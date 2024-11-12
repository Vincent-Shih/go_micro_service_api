# Cus Error

## Usage

Create a CusErr

``` go
func foo() *cus_err.CusError {
    // Case 1.
   cusErr := cus_err.New(cus_err.InternalServerError, "Your error message")

   // Case 2: New a cus_err with other error
   otherErr := someService()
   cusErr = cus_err.New(cus_err.InternalServerError, "Your error message",otherErr)

   return cusErr
}
```

Compare with cusCode

``` go
func foo() {
    cusErr := cus_err.New(cus_err.InternalServerError, "Your error message")
    if cusErr.Code() == cus_err.InternalServerError {
        // Do something...
    }
}
```

Log a cusError 

``` go
func foo(ctx context.Context) {
    cusErr := cus_err.New(cus_err.InternalServerError, "Your error message")

    // If you want some field to log also, use `NewField()` to record with your err
    cus_otel.Error(ctx, err.Message(), cus_otel.NewField("token", req.AccessToken))

    // Or you can just send log simply
    cus_otel.Error(ctx, err.Error())
}
```
