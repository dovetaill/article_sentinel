package bootstrap

// BuildServerRuntime 组装 server 入口运行所需的共享资源。
func BuildServerRuntime(configPath string) (*Runtime, error) {
	return buildRuntime(configPath)
}
