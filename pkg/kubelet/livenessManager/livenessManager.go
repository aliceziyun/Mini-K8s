package livenessManager

import (
	"Mini-K8s/pkg/object"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// PodLifecycleEventType define the event type of pod life cycle events.
type PodLifecycleEventType string

type LivenessManager struct {
}

func (*LivenessManager) Start() {

}

const (
	eventChannelSize = 10
	relistPeriod     = 10 * time.Second
)

// var log = logger.Log("PLEG")

const (
	// ContainerStarted - event type when the new state of container is running.
	ContainerStarted PodLifecycleEventType = "ContainerStarted"
	// ContainerDied - event type when the new state of container is exited.
	ContainerDied PodLifecycleEventType = "ContainerDied"
	// ContainerRemoved - event type when the old state of container is exited.
	ContainerRemoved PodLifecycleEventType = "ContainerRemoved"
	// ContainerNeedStart - event type when the container is needed to start.
	ContainerNeedStart PodLifecycleEventType = "ContainerNeedStart"
	// ContainerNeedRestart - event type when the container needs to restart.
	ContainerNeedRestart PodLifecycleEventType = "ContainerNeedRestart"
	// ContainerNeedCreateAndStart - event type when the container needs to create and start.
	ContainerNeedCreateAndStart PodLifecycleEventType = "ContainerNeedCreateAndStart"
	// ContainerNeedRemove - event type when the container needs to be removed.
	ContainerNeedRemove PodLifecycleEventType = "ContainerNeedRemove"
	// PodSync is used to trigger syncing of a pod when the observed change of
	// the state of the pod cannot be captured by any single event above.
	PodSync PodLifecycleEventType = "PodSync"
	// ContainerChanged - event type when the new state of container is unknown.
	ContainerChanged PodLifecycleEventType = "ContainerChanged"
)

// PodLifecycleEvent is an event that reflects the change of the pod state.
type PodLifecycleEvent struct {
	// The pod ID.
	ID string
	// The api object pod itself Pod
	Pod *object.Pod
	// The type of the event.
	Type PodLifecycleEventType
	// The accompanied data which varies based on the event type.
	//   - ContainerStarted/ContainerStopped: the container name (string).
	//   - All other event types: unused.
	Data interface{}
}

const (
	// StateCreated indicates a container that has been created (e.g. with docker create) but not started.
	StateCreated string = "created"
	// StateRunning indicates a currently running container.
	StateRunning string = "running"
	// StateExited indicates a container that ran and completed ("stopped" in other contexts, although a created container is technically also "stopped").
	StateExited string = "exited"
	// StateUnknown encompasses all the states that we currently don't care about (like restarting, paused, dead).
	StateUnknown string = "unknown"
)

type ContainerStatus struct {
	// ID of the container.
	ID string
	// Name of the container.
	Name string
	// Status of the container.(created,running,exited,pending?)
	State string
	// Creation time of the container.
	CreatedAt time.Time
	// Start time of the container.
	StartedAt time.Time
	// Finish time of the container.
	FinishedAt time.Time
	// Exit code of the container.
	ExitCode int
	// ID of the image.
	ImageID string
	// Number of times that the container has been restarted.
	RestartCount int
	// A string stands for the error
	Error string
	// The status of resource usage
	// ResourcesUsage ResourcesUsage
	// PortBindings
	// PortBindings PortBindings
}

// PodStatus represents the status of the pod and its containers.
type PodStatus struct {
	// ID of the pod.
	ID string
	// Name of the pod.
	Name string
	// Namespace of the pod.
	Namespace string
	// All IPs assigned to this pod
	IPs []string
	// PodLifecycle of containers in the pod.
	ContainerStatuses []*ContainerStatus
	// PortBindings      container.PortBindings
}
type PodStatuses = map[string]*PodStatus

func (podStatus *PodStatus) GetContainerStatusByName(name string) *ContainerStatus {
	for _, cs := range podStatus.ContainerStatuses {
		if cs.Name == name {
			return cs
		}
	}
	return nil
}

type podStatusRecord struct {
	OldStatus     *PodStatus
	CurrentStatus *PodStatus
}

type podStatusRecords map[string]*podStatusRecord

func (statusRecords podStatusRecords) UpdateRecord(podUID string, newStatus *PodStatus) {
	if record, exists := statusRecords[podUID]; exists {
		record.OldStatus = record.CurrentStatus
		record.CurrentStatus = newStatus
	} else {
		statusRecords[podUID] = &podStatusRecord{
			OldStatus:     nil,
			CurrentStatus: newStatus,
		}
	}
}

func (statusRecords podStatusRecords) RemoveRecord(podUID string) {
	delete(statusRecords, podUID)
}

func (statusRecords podStatusRecords) GetRecord(podUID string) *podStatusRecord {
	return statusRecords[podUID]
}

type PodRestartContainerArgs struct {
	ContainerID       string
	ContainerFullName string
}

type Manager interface {
	Updates() chan *PodLifecycleEvent
	Start()
}

type OptFn interface {
	GetPod(podUID string) *object.Pod
	AddPod(podUID string, pod *object.Pod)
	UpdatePod(podUID string, newPod *object.Pod)
	DeletePod(podUID string)
	GetPodStatuses() PodStatuses
	Start()
}

func NewPlegManager(statusManager OptFn) Manager {
	return &manager{
		eventCh:          make(chan *PodLifecycleEvent, eventChannelSize),
		statusManager:    statusManager,
		podStatusRecords: make(podStatusRecords),
	}
}

type manager struct {
	eventCh          chan *PodLifecycleEvent
	statusManager    OptFn
	podStatusRecords podStatusRecords
}

func newPodLifecycleEvent(podUID string, pod *object.Pod, eventType PodLifecycleEventType, data interface{}) *PodLifecycleEvent {
	return &PodLifecycleEvent{
		ID:   podUID,
		Pod:  pod,
		Type: eventType,
		Data: data,
	}
}

func (m *manager) addStartedLifecycleEvent(podUID string, pod *object.Pod, containerID string) {
	m.eventCh <- newPodLifecycleEvent(podUID, pod, ContainerStarted, containerID)
}

func (m *manager) addNeedRemoveLifecycleEvent(podUID string, pod *object.Pod, containerID string) {
	m.eventCh <- newPodLifecycleEvent(podUID, pod, ContainerNeedRemove, containerID)
}

func (m *manager) addNeedRestartLifecycleEvent(podUID string, pod *object.Pod, args PodRestartContainerArgs) {
	m.eventCh <- newPodLifecycleEvent(podUID, pod, ContainerNeedRestart, args)
}

func (m *manager) addNeedStartLifecycleEvent(podUID string, pod *object.Pod, args PodRestartContainerArgs) {
	m.eventCh <- newPodLifecycleEvent(podUID, pod, ContainerNeedStart, args)
}

func (m *manager) addNeedCreateAndStartLifecycleEvent(podUID string, pod *object.Pod, target *object.Container) {
	m.eventCh <- newPodLifecycleEvent(podUID, pod, ContainerNeedCreateAndStart, target)
}

func (m *manager) addDiedLifecycleEvent(podUID string, pod *object.Pod, containerID string) {
	m.eventCh <- newPodLifecycleEvent(podUID, pod, ContainerDied, containerID)
}

func (m *manager) addRemovedLifecycleEvent(podUID string, pod *object.Pod, containerID string) {
	m.eventCh <- newPodLifecycleEvent(podUID, pod, ContainerRemoved, containerID)
}

func (m *manager) addPodSyncLifecycleEvent(podUID string, pod *object.Pod, containerID string) {
	m.eventCh <- newPodLifecycleEvent(podUID, pod, PodSync, containerID)
}

func (m *manager) addChangedLifecycleEvent(podUID string, pod *object.Pod, containerID string) {
	m.eventCh <- newPodLifecycleEvent(podUID, pod, ContainerChanged, containerID)
}

func (m *manager) removeAllContainers(runtimePodStatus *PodStatus) {
	for _, cs := range runtimePodStatus.ContainerStatuses {
		m.addNeedRemoveLifecycleEvent(runtimePodStatus.ID, nil, cs.ID)
	}
}

func ParseContainerFullName(containerFullName string) (succ bool, containerName, podName, podNamespace string, podUID string, restartCount int) {
	if containerFullName[0] == '/' {
		containerFullName = containerFullName[1:]
	}
	tokens := strings.Split(containerFullName, "_")
	var err error
	succ = false
	if numTokens := len(tokens); numTokens == 6 {
		succ = true
		containerName, podName, podNamespace, podUID = tokens[1], tokens[2], tokens[3], tokens[4]
		restartCount, err = strconv.Atoi(tokens[5])
		succ = err == nil
	}
	return
}

// compareAndProduceLifecycleEvents compares given runtime pod statuses
// with pod api object, and produce corresponding lifecycle events
// / TODO what about pause?
func (m *manager) compareAndProduceLifecycleEvents(pod *object.Pod, runtimePodStatus *PodStatus) {
	podUID := runtimePodStatus.ID
	m.podStatusRecords.UpdateRecord(podUID, runtimePodStatus)
	record := m.podStatusRecords.GetRecord(podUID)
	oldStatus, currentStatus := record.OldStatus, record.CurrentStatus

	// apiPod == nil means the pod is no longer existent, skip this
	if pod == nil {
		return
	}

	notIncludedContainerNameMap := make(map[string]struct{})
	for _, c := range pod.Containers() {
		notIncludedContainerNameMap[c.Name] = struct{}{}
	}

	for _, cs := range currentStatus.ContainerStatuses {
		parseSucc, containerName, _, _, _, _ := ParseContainerFullName(cs.Name)
		// illegal containerName, need remove it
		if !parseSucc {
			//m.addNeedRemoveLifecycleEvent(podUID, cs.ID)
			continue
		}

		// Only deal with it when state has changed
		needDealWith := oldStatus == nil
		if !needDealWith {
			oldCs := oldStatus.GetContainerStatusByName(cs.Name)
			needDealWith = oldCs == nil || oldCs.State != cs.State
		}

		if needDealWith {
			switch cs.State {
			case StateRunning:
			case StateCreated:
				//if apiPod.GetContainerByName(containerName) == nil {
				//	m.addNeedRemoveLifecycleEvent(podUID, cs.ID)
				//}
				break
			// Need restart it
			case StateExited:
				if pod.GetContainerByName(containerName) != nil {
					m.addNeedRestartLifecycleEvent(podUID, pod, PodRestartContainerArgs{cs.ID, cs.Name})
				}
			default:
				m.addChangedLifecycleEvent(podUID, pod, cs.ID)
			}
		}
		delete(notIncludedContainerNameMap, containerName)
	}
	// Need to create all the container that has not been created
	for notIncludeContainerName := range notIncludedContainerNameMap {
		m.addNeedCreateAndStartLifecycleEvent(podUID, pod, pod.GetContainerByName(notIncludeContainerName))
	}
}

func (m *manager) relist() error {
	// Step 1: Get all *runtime* pod statuses
	// If there are no available pod info, just return
	runtimePodStatuses := m.statusManager.GetPodStatuses()
	if runtimePodStatuses == nil {
		return nil
	}

	// Step 2: Get pod api object, and according to the api object, produce lifecycle events
	var apiPod *object.Pod
	for podUID, runtimePodStatus := range runtimePodStatuses {
		apiPod = m.statusManager.GetPod(podUID)
		m.compareAndProduceLifecycleEvents(apiPod, runtimePodStatus)
	}

	return nil
}

func Period(delay time.Duration, period time.Duration, handler func()) {
	<-time.After(delay)
	tick := time.Tick(period)
	for {
		handler()
		<-tick
	}
}
func After(d time.Duration, handler func()) {
	<-time.After(d)
	handler()
}

func (m *manager) run() {
	Period(relistPeriod, relistPeriod, func() {
		if err := m.relist(); err != nil {
			fmt.Println(err)
		}
	})
}

func (m *manager) Updates() chan *PodLifecycleEvent {
	return m.eventCh
}

func (m *manager) Start() {
	go m.run()
}
