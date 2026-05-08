package actions

type BatchActionInput struct {
	OrgID        uint64
	TaskID       uint64
	ResultIDs    []uint64
	OperatorID   uint64
	OperatorName string
	Reason       string
}

type BatchActionSummary struct {
	ActionID     uint64 `json:"action_id"`
	TargetCount  int64  `json:"target_count"`
	SuccessCount int64  `json:"success_count"`
	FailCount    int64  `json:"fail_count"`
	SkipCount    int64  `json:"skip_count"`
	Status       string `json:"status"`
	ActionType   string `json:"action_type"`
}
