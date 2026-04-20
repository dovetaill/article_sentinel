package bootstrap

// BuildSchedulerRuntime 组装 scheduler 入口运行所需的共享资源。
func BuildSchedulerRuntime(configPath string) (*Runtime, error) {
	return buildRuntime(configPath)
}
