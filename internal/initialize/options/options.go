package options

// Options 定义全局配置选项
type Options struct {
	Hosts       string
	Ports       string
	Concurrency int
	Timeout     int
	Verbose     bool
	Mode        string
	OutputFile  string
	LogLevel    string
}
