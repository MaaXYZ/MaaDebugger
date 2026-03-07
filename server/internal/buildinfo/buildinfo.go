package buildinfo

import (
	"fmt"
	"time"
)

// 通过 -ldflags 在编译时注入相关信息
var (
	Version   string = "dev"
	CommitSHA string = "dev"
	BuildTime string = fmt.Sprintf("%d", time.Now().Unix())
)
