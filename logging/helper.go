package logging

var (
	debugEnabled   = false
	verboseEnabled = false
)

func Setup(debug, verbose bool) {
	debugEnabled = debug || verbose
	verboseEnabled = verbose
}

func IsDebugEnabled() bool {
	return debugEnabled
}

func IsVerboseEnabled() bool {
	return verboseEnabled
}
