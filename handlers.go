package log

import (
	"errors"
	"github.com/garyburd/redigo/redis"
	"golang.org/x/crypto/ssh/terminal"
	"gopkg.in/inconshreveable/log15.v2"
	"io"
	"os"
)

var BadConf error = errors.New("Bad configuration")

type FormatConf interface{}

// MakeFormatter constructs a object of type Format
// based on the specified format conf and returns it
// Currently @format has to be a string with one of
// json | json_pretty | logfmt | terminal
func MakeFormatter(format FormatConf) (Format, error) {

	format_name, ok := format.(string)
	if !ok {
		return nil, BadConf
	}

	switch format_name {

	case "json":
		return log15.JsonFormat(), nil

	case "json_pretty":
		return log15.JsonFormatEx(true, true), nil

	case "logfmt":
		return log15.LogfmtFormat(), nil

	case "terminal":
		return log15.TerminalFormat(), nil

	}

	return nil, BadConf
}

type HandlerConf []interface{}

// MakeHandler accepts a handler configuration and constructs a usable
//  handler. The handler config can be used to create sophistacted behavior
//  by composing handlers as described here
//  "https://godoc.org/gopkg.in/inconshreveable/log15.v2#Handler"
//
//  A HandlerConf is a sequence of values of the following format
//		[handlerType string, args ...interface{}]
//
//      eg: HandlerConf{"file", "/tmp/test.log", "json"}
//		In the above example, we have defined a file handler where the logs
//		are to be written to the file "/tmp/test.log" in "json" format.
//
//		Here is an example of a slightly more complex handler
//		eg: HandlerConf{"level_filter", "debug",
//				HandlerConf{"file", "/tmp/test.log", "json"},
//			}
//		In the above example, we defined a file handler just like above
//		but in wrapping it with a "level_filter" handler, we've specified
//		that only log statements that are greater than or equal to the
//		specified level "debug".
//
//	NOTE: for additional information about the following please
//	refer to the godoc link placed above.
//
//	List of handler formats:
//		json
//      json_pretty
//		logfmt
//		terminal
//
//	List of handlers:
//
//	- buffered (bufSize int, handler HandlerConf)
//	- caller_file (handler HandlerConf)
//	- caller_func (handler HandlerConf)
//	- caller_stack (format string, handler HandlerConf)
//	- discard ()
//	- failover (handler ...HandlerConf)
//  - file (path string, format string)
//  - lazy (handler HandlerConf)
//  - level_filter (level string, handler HandlerConf)
//		level = debug | info | warn | error | crit
//  - match_filter (key string, value string|int|float, handler HandlerConf)
//	- multi (handler ...HandlerConf)
//	- net (network string, address string, format string)
//	- stream (stream string, format string)
//		stream = stdout | stderr
//	- sync (handler HandlerConf)
//	- syslog (tag string, format string)
//	- syslog_net (net string, address string, tag string, format string)
//	- redis (ip_port string, channel string)
//		`ip_port` is of the format "ip:port". port part is optional. on omission
//			the default redis port 6379 is assumed.
//		`channel` is the name of the redis channel to which the log statements
//			are to be written.
func MakeHandler(conf HandlerConf) (Handler, error) {
	if len(conf) < 1 {
		return nil, BadConf
	}

	name := conf[0].(string)
	args := conf[1:]

	switch name {

	case "buffered":
		// buffered (bufSize int, handler HandlerConf)

		if len(args) != 2 {
			return nil, BadConf
		}

		bufSize, ok := args[0].(int)
		if !ok {
			return nil, BadConf
		}

		hdata, ok := args[1].(HandlerConf)
		if !ok {
			return nil, BadConf
		}

		h, err := MakeHandler(hdata)
		if err != nil {
			return nil, err
		}

		return log15.BufferedHandler(bufSize, h), nil

	case "caller_file":
		// caller_file (handler HandlerConf)

		if len(args) != 1 {
			return nil, BadConf
		}

		hdata, ok := args[0].(HandlerConf)
		if !ok {
			return nil, BadConf
		}

		h, err := MakeHandler(hdata)
		if err != nil {
			return nil, err
		}

		return log15.CallerFileHandler(h), nil

	case "caller_func":
		// caller_func (handler HandlerConf)

		if len(args) != 1 {
			return nil, BadConf
		}

		hdata, ok := args[0].(HandlerConf)
		if !ok {
			return nil, BadConf
		}

		h, err := MakeHandler(hdata)
		if err != nil {
			return nil, err
		}

		return log15.CallerFuncHandler(h), nil

	case "caller_stack":
		// caller_stack (format string, handler HandlerConf)

		if len(args) != 2 {
			return nil, BadConf
		}

		format, ok := args[0].(string)
		if !ok {
			return nil, BadConf
		}

		hdata, ok := args[1].(HandlerConf)
		if !ok {
			return nil, BadConf
		}

		h, err := MakeHandler(hdata)
		if err != nil {
			return nil, err
		}

		return log15.CallerStackHandler(format, h), nil

	case "discard":
		// discard ()

		if len(args) != 0 {
			return nil, BadConf
		}

		return log15.DiscardHandler(), nil

	case "failover":
		// failover (handler ...HandlerConf)

		hs := make([]log15.Handler, len(args))
		for i := 0; i < len(args); i++ {
			h, err := MakeHandler(args[i].(HandlerConf))
			if err != nil {
				return nil, err
			}
			hs = append(hs, h)
		}

		return log15.FailoverHandler(hs...), nil

	case "file":
		// file (path string, format string)

		if len(args) != 2 {
			return nil, BadConf
		}

		path, ok := args[0].(string)
		if !ok {
			return nil, BadConf
		}

		formatter, err := MakeFormatter(args[1])
		if err != nil {
			return nil, err
		}

		return log15.FileHandler(path, formatter)

	case "lazy":
		// lazy (handler HandlerConf)

		if len(args) != 1 {
			return nil, BadConf
		}

		hdata, ok := args[0].(HandlerConf)
		if !ok {
			return nil, BadConf
		}

		h, err := MakeHandler(hdata)
		if err != nil {
			return nil, err
		}

		return log15.LazyHandler(h), nil

	case "level_filter":
		// level_filter (level string, handler HandlerConf)
		//		level = debug | info | warn | error | crit

		if len(args) != 2 {
			return nil, BadConf
		}

		lvlString, ok := args[0].(string)
		if !ok {
			return nil, BadConf
		}

		lvl, err := log15.LvlFromString(lvlString)
		if err != nil {
			return nil, BadConf
		}

		hdata, ok := args[1].(HandlerConf)
		if !ok {
			return nil, BadConf
		}

		h, err := MakeHandler(hdata)
		if err != nil {
			return nil, err
		}

		return log15.LvlFilterHandler(lvl, h), nil

	case "match_filter":
		// match_filter (key string, value string|int|float, handler HandlerConf)

		if len(args) != 3 {
			return nil, BadConf
		}

		key, ok := args[0].(string)
		if !ok {
			return nil, BadConf
		}

		value, ok := args[1].(interface{})
		if !ok {
			return nil, BadConf
		}

		hdata, ok := args[2].(HandlerConf)
		if !ok {
			return nil, BadConf
		}

		h, err := MakeHandler(hdata)
		if err != nil {
			return nil, err
		}

		return log15.MatchFilterHandler(key, value, h), nil

	case "multi":
		// multi (handler ...HandlerConf)

		hs := make([]log15.Handler, len(args))
		for i := 0; i < len(args); i++ {
			h, err := MakeHandler(args[i].(HandlerConf))
			if err != nil {
				return nil, err
			}
			hs[i] = h
		}

		return log15.MultiHandler(hs...), nil

	case "net":
		// net (network string, address string, format string)

		if len(args) != 3 {
			return nil, BadConf
		}

		network, ok := args[0].(string)
		if !ok {
			return nil, BadConf
		}

		address, ok := args[1].(string)
		if !ok {
			return nil, BadConf
		}

		formatter, err := MakeFormatter(args[2])
		if err != nil {
			return nil, err
		}

		return log15.NetHandler(network, address, formatter)

	case "stream":
		// stream (stream string, format string)
		//		stream = stdout | stderr

		if len(args) != 2 {
			return nil, BadConf
		}

		stream_name, ok := args[0].(string)
		if !ok {
			return nil, BadConf
		}

		var stream io.Writer

		switch stream_name {
		case "stdout":
			stream = os.Stdout
		case "stderr", "":
			stream = os.Stderr
		default:
			return nil, BadConf
		}

		formatter, err := MakeFormatter(args[1])
		if err != nil {
			return nil, err
		}

		return log15.StreamHandler(stream, formatter), nil

	case "sync":
		// sync (handler HandlerConf)

		if len(args) != 1 {
			return nil, BadConf
		}

		hdata, ok := args[0].(HandlerConf)
		if !ok {
			return nil, BadConf
		}

		h, err := MakeHandler(hdata)
		if err != nil {
			return nil, err
		}

		return log15.SyncHandler(h), nil

	case "syslog":
		// syslog (tag string, format string)

		if len(args) != 2 {
			return nil, BadConf
		}

		tag, ok := args[0].(string)
		if !ok {
			return nil, BadConf
		}

		formatter, err := MakeFormatter(args[1])
		if err != nil {
			return nil, err
		}

		return log15.SyslogHandler(tag, formatter)

	case "syslog_net":
		// syslog_net (net string, address string, tag string, format string)

		if len(args) != 4 {
			return nil, BadConf
		}

		network, ok := args[0].(string)
		if !ok {
			return nil, BadConf
		}

		address, ok := args[1].(string)
		if !ok {
			return nil, BadConf
		}

		tag, ok := args[2].(string)
		if !ok {
			return nil, BadConf
		}

		formatter, err := MakeFormatter(args[3])
		if err != nil {
			return nil, err
		}

		return log15.SyslogNetHandler(network, address, tag, formatter)

	case "redis":
		// redis (ip_port string, channel string)

		if len(args) != 2 {
			return nil, BadConf
		}

		ip_port, ok := args[0].(string)
		if !ok {
			return nil, BadConf
		}

		channel, ok := args[1].(string)
		if !ok {
			return nil, BadConf
		}

		redis_h := RedisHandler{Loc: ip_port, Channel: channel}
		err := redis_h.Init()
		if err != nil {
			return nil, err
		}

		return log15.SyncHandler(redis_h), nil

	default:
		return nil, BadConf
	}

}

type RedisHandler struct {
	Loc     string
	Channel string

	conn      redis.Conn
	formatter log15.Format
}

func (p *RedisHandler) Init() error {
	var err error
	p.formatter = log15.JsonFormat()
	p.conn, err = redis.Dial("tcp", p.Loc)
	return err
}

func (p RedisHandler) Log(r *log15.Record) error {
	b := p.formatter.Format(r)
	_, err := p.conn.Do("PUBLISH", p.Channel, b)
	return err
}

// MakeBasicHandler prepares a log handler that writes to both
// a file at @fpath and stderr both with log level @lvl
// if @fpath is "", then it assumes it shouldn't write to a file
// and if @quiet is true, then it doesn't print to stderr
func MakeBasicHandler(fpath, lvl string, quiet bool) (Handler, error) {
	var (
		fileHandlerConf   HandlerConf = HandlerConf{"discard"}
		streamHandlerConf HandlerConf = HandlerConf{"discard"}
	)
	if fpath != "" {
		fileHandlerConf = HandlerConf{"file", fpath, "json"}
	}
	if !quiet {
		if terminal.IsTerminal(int(os.Stdout.Fd())) {
			streamHandlerConf = HandlerConf{"stream", "stderr", "terminal"}
		} else {
			streamHandlerConf = HandlerConf{"stream", "stderr", "json"}
		}
	}

	finalHandlerConf := HandlerConf{
		"level_filter", lvl, HandlerConf{
			"caller_file", HandlerConf{
				"multi", fileHandlerConf, streamHandlerConf,
			},
		},
	}
	return MakeHandler(finalHandlerConf)
}

// MakeNoopHandler prepares a log handler that discards all the log
// statements that are sent to it
func MakeNoopHandler() (Handler, error) {
	return MakeHandler(HandlerConf{"discard"})
}
