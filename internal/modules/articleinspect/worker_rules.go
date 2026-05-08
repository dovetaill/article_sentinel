package articleinspect

import workerpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/worker"

func decodeTaskRules(snapshot string) ([]KeywordRule, error) {
	return workerpkg.DecodeTaskRules(snapshot)
}

func parseArticleStateFilter(value string) int8 {
	return workerpkg.ParseArticleStateFilter(value)
}

func resolveTaskStatus(totalScanned, failCount int64) string {
	return workerpkg.ResolveTaskStatus(totalScanned, failCount)
}
