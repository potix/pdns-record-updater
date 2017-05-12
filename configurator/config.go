package configurator

// Target is config of target
type Target struct {
	Name        string
	TargetType  string  `json:"type" toml:"type" yaml:"type"`
	Destination string
	Status      string
	Regex       string
	Retry       uint32
	RetryWait   uint32
	Timeout     uint32
	alive       bool
}

// Record is config of record
type Record struct {
	Name                 string
	RecordType           string `json:"type" toml:"type" yaml:"type"`
	Content              string
        Targets              []*target
	EvalRule             string
	WatchInterval        uint32
	currentIntervalCount uint32
	progress             uint32
	alive                bool
}

// Config is config
type Config struct {
	Records       []*record
	ConfigLoggers *belog.ConfigLoggers
}


//---------------------------------------------------------------------------
//"records" : [
//    {
//       "name": 2.2.2.2
//       "type": 2.2.2.2
//       "content": 2.2.2.2
//       "targets"[
//          {a, tcp_conn 2.2.2.2:80},
//          {b, http_status 2.2.2.2:80},
//          {c, https_status 2.2.2.2:80},
//          {d, tcp_regex 2.2.2.2:80},
//          {e, http_status_regex 2.2.2.2:80},
//          {f, https_status_regex 2.2.2.2:80},
//        {g, icmp 3.3.3.3.3},
//       ]
//       "eval" :  "a & (b & c) & d | (c|f)"
//    },
//],
//loggers:
//  test1:
//    filter:
//      structname: LogLevelFilter
//      structsetters:
//      - settername: SetLogLevel
//        setterparams:
//        - "8"
//    formatter:
//      structname: StandardFormatter
//      structsetters:
//      - settername: SetDateTimeLayout
//        setterparams:
//        - 2006-01-02 15:04:05 -0700 MST
//      - settername: SetLayout
//        setterparams:
//        - '%(dateTime) [%(logLevel)] (%(pid)) %(programCounter) %(loggerName) %(fileName) %(lineNum) %(message)'
//    handlers:
//    - structname: ConsoleHandler
//      structsetters:
//      - settername: SetOutputType
//        setterparams:
//        - "2"
//    - structname: SyslogHandler
//      structsetters:
//      - settername: SetNetworkAndAddr
//        setterparams:
//        - ""
//        - ""
//      - settername: SetTag
//        setterparams:
//        - "test"
//      - settername: SetFacility
//        setterparams:
//        - "Daemon"
//    - structname: RotationFileHandler
//      structsetters:
//      - settername: SetLogFileName
//        setterparams:
//        - belog-test.log
//      - settername: SetLogDirPath
//        setterparams:
//        - /var/tmp/belog-test
//      - settername: SetMaxAge
//        setterparams:
////        - "3"
//      - settername: SetMaxSize
//        setterparams:
//        - "65535"
//      - settername: SetAsync
//        setterparams:
//        - "true"
//      - settername: SetAsyncFlushInterval
//        setterparams:
//        - "2"
//      - settername: SetBufferSize
//        setterparams:
//        - "1024"
//  test2:
//    filter:
//      structname: LogLevelFilter
//      structsetters: []
//    formatter:
//      structname: StandardFormatter
//      structsetters: []
//    handlers:
//    - structname: ConsoleHandler
//      structsetters: []
//    - structname: SyslogHandler
//      structsetters: []
//    - structname: RotationFileHandler
//      structsetters: []
