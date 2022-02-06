package debug

import "os"

func IsDebug() bool {
	return len(os.Args) > 1 && os.Args[1] == "debug"
}
