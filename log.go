// log package enables simple levelled logging
// This is a thin wrapper around log15 which provides most of the
// functionality. We are wrapping primarily to not have to use "log15" as
// an import (which is just weird).
package log

import (
	"fmt"
	"gopkg.in/inconshreveable/log15.v2"
	"log"
)

type Logger log15.Logger
type Ctx log15.Ctx
type Format log15.Format
type Handler log15.Handler
type Lazy log15.Lazy
type Lvl log15.Lvl
type Record log15.Record
type RecordKeyNames log15.RecordKeyNames

var New = log15.New
var Root = log15.Root

var Crit = log15.Crit
var Debug = log15.Debug
var Error = log15.Error
var Info = log15.Info
var Warn = log15.Warn

func Printf(format string, v ...interface{}) {
	Debug(fmt.Sprintf(format, v...))
}

func Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	Error(s)
	panic(s)
}

func SetHandler(hdlr Handler) {
	Root().SetHandler(hdlr)
}

type LogToLog15 struct {
}

func (p *LogToLog15) Write(b []byte) (n int, err error) {
	Info(string(b))
	return len(b), nil
}

func init() {
	// Route all logs sent to golang's built-in logger to us
	log.SetOutput(&LogToLog15{})

	hdlr, _ := MakeHandler(HandlerConf{"stream", "stderr", "terminal"})
	SetHandler(hdlr)
}
