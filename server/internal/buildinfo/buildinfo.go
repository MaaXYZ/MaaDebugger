package buildinfo

// 通过 -ldflags 在编译时注入相关信息
var (
	Version   string = "dev"
	CommitSHA string = "dev"
	BuildTime string = "0"
)
