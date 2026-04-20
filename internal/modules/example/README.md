# Starter Module Flow

`internal/modules/post` 是 starter 的官方可复制模板。标准替换流程如下。

1. 运行 `scripts/new-module.sh`
   - 示例：`bash scripts/new-module.sh article`
2. 重命名并调整模型/仓储/服务/处理器细节
   - 检查 `model.go`、`repository.go`、`service.go`、`handler.go` 中的命名与领域语义
3. 在 `internal/api/register/router.go` 注册新模块
   - 保持 router 只做 wiring，不放业务逻辑
4. 在 `internal/app/bootstrap/schema.go` 注册模型（如果当前 schema sync 仍依赖模型注册）
5. 扩展测试
   - 先改模块测试，再补充路由装配测试，确保新模块契约可回归

## Notes

- `scripts/new-module.sh` 只做“明显 token”替换，不会自动完成你的业务语义迁移。
- 若你的模块命名是复合词，请重点复查自动替换后的类型名和路由词形。
