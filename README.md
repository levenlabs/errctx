# errctx

The errctx package allows for setting and retrieving contextual information using
`error` objects.

## Usage
```
err := errors.New("something bad happened")
// set the userID that caused this error
return errctx.Set(err, "userID", userID)
```

Later on you can get the userID back:
```
userID, ok := errctx.Get(err, "userID").(int64)
```

If you want to check the original error object:
```
if errors.Is(err, mgo.ErrNotFound) {

}
```

Additionally, if you want to store the source of the original error you can
use `Mark` and `Line`:
```
err := errors.New("something bad happened")
// store the filename and line number where Mark was called
return errctx.Mark(err)
```

Later you can use `Line` to get the filename:line where Mark was first called:
```
fileLine, ok := errctx.Line(err)
```
