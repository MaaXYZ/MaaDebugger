package buildinfo

import (
	"fmt"
	"time"
)

// 通过 -ldflags 在编译时注入相关信息
var (
	Version   string = "dev"
	CommitSHA string = "dev"
	BuildTime string = fmt.Sprintf("%s", time.Now().Format("2006-01-02 15:04:05"))
)
