package deployment

import (
	"Mini-K8s/pkg/listwatcher"
	"Mini-K8s/third_party/queue"
	"time"
)

// DeploymentController :负责同步真实运行的Pod和存储在系统中期望的Pod数量
type DeploymentController struct {
	ls *listwatcher.ListWatcher //deployment informer + pod infomer

	enqueueDeployment func(deployment *apps.Deployment)

	resyncInterval time.Duration
	stopChannel    chan struct{}
	apiServerBase  string

	queue queue.ConcurrentQueue //存储deployment的队列
}
