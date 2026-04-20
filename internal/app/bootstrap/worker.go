package bootstrap

// BuildWorkerRuntime 组装 worker 入口运行所需的共享资源。
func BuildWorkerRuntime(configPath string) (*Runtime, error) {
	return buildRuntime(configPath)
}
