package articleinspect

import (
	taskspkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/tasks"
	"gorm.io/gorm"
)

var ErrInvalidTaskInput = taskspkg.ErrInvalidTaskInput
var ErrTaskNotFound = taskspkg.ErrTaskNotFound
var ErrTaskDeleteNotAllowed = taskspkg.ErrTaskDeleteNotAllowed

type TaskService = taskspkg.TaskService

func NewTaskService(db *gorm.DB, keywords *KeywordRepository, articles *ArticleRepository) *TaskService {
	return taskspkg.NewTaskService(db, keywords, articles)
}
