package configsync

import (
	"fmt"
	"os"
)

func stderr(format string, args ...interface{}) {
	a := fmt.Sprintf(format, args...)
	os.Stderr.WriteString(a)
	os.Stderr.WriteString("\n")
}

func stdout(format string, args ...interface{}) {
	a := fmt.Sprintf(format, args...)
	os.Stdout.WriteString(a)
	os.Stdout.WriteString("\n")
}
