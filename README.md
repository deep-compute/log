# log
light abstraction over https://github.com/inconshreveable/log15

# Usage

Initialize it once in the main program as follows - 

```go
import "github.com/deep-compute/log"

func main() {
    quiet := True
    filename := "/tmp/test.log"
    level := "debug"

    hdlr, err := log.MakeBasicHandler(filename, level, quiet)
    if err != nil {
        panic(err)
    }

    log.SetHandler(hdlr)
}
```

All other packages need simply import and start logging wherever necessary

```go
import "github.com/deep-compute/log"

func Foo() {
    log.Info("inside function Foo")
}
```
