package container

import (
	"context"
	"errors"
	"fmt"
	eventtypes "github.com/containerd/containerd/api/events"
	"github.com/containerd/containerd/v2/core/events"
	"github.com/containerd/nerdctl/v2/pkg/clientutil"
	"github.com/containerd/nerdctl/v2/pkg/containerinspector"
	"github.com/containerd/nerdctl/v2/pkg/inspecttypes/native"
	"github.com/containerd/typeurl/v2"
	"github.com/petermattis/goid"
	"golang.org/x/sys/unix"
	"k8s.io/client-go/util/workqueue"
	"log"
	"strings"
)

// Kube default sandbox annotation keys
const (
	//CRI-O
	CRIO_CONTAINER_TYPE    = "io.kubernetes.cri-o.ContainerType"
	CRIO_IMAGE_NAME        = "io.kubernetes.cri-o.ImageName"
	CRIO_SANDBOX_NAMESPACE = "io.kubernetes.pod.namespace"
	CRIO_SANDBOX_NAME      = "io.kubernetes.pod.name"
	CRIO_CONTAINER_NAME    = "io.kubernetes.container.name"
	CRIO_SANDBOX_ID        = "io.kubernetes.cri-o.SandboxID"
	CRIO_LOG_DIRECTORY     = "io.kubernetes.cri-o.LogPath"

	// CRI
	CONTAINER_TYPE    = "io.kubernetes.cri.container-type"
	SANDBOX_ID        = "io.kubernetes.cri.sandbox-id"
	SANDBOX_NAME      = "io.kubernetes.cri.sandbox-name"
	SANDBOX_NAMESPACE = "io.kubernetes.cri.sandbox-namespace"
	SANDBOX_UID       = "io.kubernetes.cri.sandbox-uid"
	LOG_DIRECTORY     = "io.kubernetes.cri.sandbox-log-directory"

	// Kube container only annotation keys
	CONTAINER_NAME = "io.kubernetes.cri.container-name"
	IMAGE_NAME     = "io.kubernetes.cri.image-name"
)

const CONTAINER_TYPE_CONTAINER = "container"

const CONTAINER_TYPE_SANDBOX = "sandbox"

func NewWorkQueue(ctx context.Context) *WorkQueue {
	return &WorkQueue{
		TaskCreate:      &TaskCreateQ{NewEventQ(ctx)},
		TaskExecAdded:   &TaskExecAddedQ{NewEventQ(ctx)},
		TaskExecStarted: &TaskExecStartedQ{NewEventQ(ctx)},
		TaskDelete:      &TaskDeleteQ{NewEventQ(ctx)},
		TaskExit:        &TaskExitQ{NewEventQ(ctx)},
		TaskPaused:      &TaskPausedQ{NewEventQ(ctx)},
		TaskResumed:     &TaskResumedQ{NewEventQ(ctx)},
		TaskOOM:         &TaskOOMQ{NewEventQ(ctx)},

		ContainerCreate:         &ContainerCreateQ{NewEventQ(ctx)},
		ContainerCreate_Runtime: &ContainerCreate_RuntimeQ{NewEventQ(ctx)},
		ContainerDelete:         &ContainerDeleteQ{NewEventQ(ctx)},
		ContainerUpdate:         &ContainerUpdateQ{NewEventQ(ctx)},
	}
}

func (w *WorkQueue) Run() {
	go func() {
		err := w.ContainerCreate.Run()
		if err != nil {
			return
		}
	}()

	go func() {
		err := w.ContainerDelete.Run()
		if err != nil {
			return
		}
	}()

	go func() {
		err := w.ContainerCreate_Runtime.Run()
		if err != nil {
			return
		}
	}()

	go func() {
		err := w.ContainerUpdate.Run()
		if err != nil {
			return
		}
	}()
	go func() {
		err := w.TaskCreate.Run()
		if err != nil {
			return
		}
	}()
	go func() {
		err := w.TaskExit.Run()
		if err != nil {
			return
		}
	}()
	go func() {
		err := w.TaskExecAdded.Run()
		if err != nil {
			return
		}
	}()
	go func() {
		err := w.TaskExecStarted.Run()
		if err != nil {
			return
		}
	}()
	go func() {
		err := w.TaskOOM.Run()
		if err != nil {
			return
		}
	}()

	go func() {
		err := w.TaskResumed.Run()
		if err != nil {
			return
		}
	}()
	go func() {
		err := w.TaskPaused.Run()
		if err != nil {
			return
		}
	}()
	go func() {
		err := w.TaskDelete.Run()
		if err != nil {
			return
		}
	}()
}

type WorkQueue struct {
	TaskCreate              *TaskCreateQ
	TaskExecAdded           *TaskExecAddedQ
	TaskExecStarted         *TaskExecStartedQ
	TaskDelete              *TaskDeleteQ
	TaskExit                *TaskExitQ
	TaskPaused              *TaskPausedQ
	TaskResumed             *TaskResumedQ
	TaskOOM                 *TaskOOMQ
	ContainerCreate         *ContainerCreateQ
	ContainerCreate_Runtime *ContainerCreate_RuntimeQ
	ContainerDelete         *ContainerDeleteQ
	ContainerUpdate         *ContainerUpdateQ
}

type WorkQ interface {
	GetQueue() *workqueue.Type
	Run() error
	Get() (interface{}, bool)
	Done(interface{})
	Add(interface{})
}

type EventQ struct {
	*workqueue.Type
	ctx context.Context
}

func NewEventQ(ctx context.Context) *EventQ {
	return &EventQ{Type: workqueue.New(), ctx: ctx}
}

func (q *EventQ) GetQueue() *workqueue.Type {
	return q.Type
}

type TaskCreateQ struct {
	*EventQ
}

func (q *TaskCreateQ) Run() error {
	log.Println("线程ID:", unix.Gettid(), goid.Get())
	for {
		select {
		case <-q.EventQ.ctx.Done():
			return nil
		default:
			item, isShutdown := q.Get()
			if isShutdown {
				log.Println("队列已关闭")
				return nil
			}
			{
				defer q.Done(item)
				e := item.(*eventtypes.TaskCreate)
				log.Println("TaskCreateQ ============", e.ContainerID, e.String())
				getDetail("TaskCreateQ", e.ContainerID)
			}
		}
	}
}

type TaskExecAddedQ struct {
	*EventQ
}

func (q *TaskExecAddedQ) Run() error {
	log.Println("线程ID:", unix.Gettid(), goid.Get())
	for {
		select {
		case <-q.EventQ.ctx.Done():
			return nil
		default:
			item, isShutdown := q.Get()
			if isShutdown {
				log.Println("队列已关闭")
				return nil
			}
			{
				defer q.Done(item)
				e := item.(*eventtypes.TaskExecAdded)
				log.Println("TaskExecAddedQ ============", e.ContainerID, e.String())
				getDetail("TaskExecAddedQ", e.ContainerID)
			}
		}

	}
}

type TaskExecStartedQ struct {
	*EventQ
}

func (q *TaskExecStartedQ) Run() error {
	log.Println("线程ID:", unix.Gettid(), goid.Get())
	for {
		select {
		case <-q.EventQ.ctx.Done():
			return nil
		default:
			item, isShutdown := q.Get()
			if isShutdown {
				log.Println("队列已关闭")
				return nil
			}
			{
				defer q.Done(item)
				e := item.(*eventtypes.TaskExecStarted)
				log.Println("TaskCreateQ ============", e.ContainerID, e.String())
				getDetail("TaskCreateQ", e.ContainerID)
			}
		}
	}
}

type TaskDeleteQ struct {
	*EventQ
}

func (q *TaskDeleteQ) Run() error {
	log.Println("线程ID:", unix.Gettid(), goid.Get())
	for {
		select {
		case <-q.EventQ.ctx.Done():
			return nil
		default:
			item, isShutdown := q.Get()
			if isShutdown {
				log.Println("队列已关闭")
				return nil
			}
			{
				defer q.Done(item)
				e := item.(*eventtypes.TaskDelete)
				log.Println("TaskDeleteQ ============", e.ID, e.String())
				getDetail("TaskDeleteQ", e.ContainerID)
			}
		}
	}
}

type TaskExitQ struct {
	*EventQ
}

func (q *TaskExitQ) Run() error {
	log.Println("线程ID:", unix.Gettid(), goid.Get())
	for {
		select {
		case <-q.EventQ.ctx.Done():
			return nil
		default:
			item, isShutdown := q.Get()
			if isShutdown {
				log.Println("队列已关闭")
				return nil
			}
			{
				defer q.Done(item)
				e := item.(*eventtypes.TaskExit)
				log.Println("TaskExitQ ============", e.ID, e.String())
				getDetail("TaskExitQ", e.ContainerID)
			}
		}
	}
}

type TaskPausedQ struct {
	*EventQ
}

func (q *TaskPausedQ) Run() error {
	log.Println("线程ID:", unix.Gettid(), goid.Get())
	for {
		select {
		case <-q.EventQ.ctx.Done():
			return nil
		default:
			item, isShutdown := q.Get()
			if isShutdown {
				log.Println("队列已关闭")
				return nil
			}
			{
				defer q.Done(item)
				e := item.(*eventtypes.TaskPaused)
				log.Println("TaskPausedQ ============", e.ContainerID, e.String())
				getDetail("TaskPausedQ", e.ContainerID)
			}
		}
	}
}

type TaskResumedQ struct {
	*EventQ
}

func (q *TaskResumedQ) Run() error {
	log.Println("线程ID:", unix.Gettid(), goid.Get())
	for {
		select {
		case <-q.EventQ.ctx.Done():
			return nil
		default:
			item, isShutdown := q.Get()
			if isShutdown {
				log.Println("队列已关闭")
				return nil
			}
			{
				defer q.Done(item)
				e := item.(*eventtypes.TaskResumed)
				log.Println("TaskResumedQ ============", e.ContainerID, e.String())
				getDetail("TaskResumedQ", e.ContainerID)
			}
		}
	}
}

type TaskOOMQ struct {
	*EventQ
}

func (q *TaskOOMQ) Run() error {
	log.Println("线程ID:", unix.Gettid(), goid.Get())
	for {
		select {
		case <-q.EventQ.ctx.Done():
			return nil
		default:
			item, isShutdown := q.Get()
			if isShutdown {
				log.Println("队列已关闭")
				return nil
			}
			{
				defer q.Done(item)
				e := item.(*eventtypes.TaskOOM)
				log.Println("TaskOOMQ ============", e.ContainerID, e.String())
				getDetail("TaskOOMQ", e.ContainerID)
			}
		}
	}
}

type ContainerCreateQ struct {
	*EventQ
}

func (q *ContainerCreateQ) Run() error {
	log.Println("线程ID:", unix.Gettid(), goid.Get())
	for {
		select {
		case <-q.EventQ.ctx.Done():
			return nil
		default:
			item, isShutdown := q.Get()
			if isShutdown {
				log.Println("队列已关闭")
				return nil
			}
			{
				defer q.Done(item)
				e := item.(*eventtypes.ContainerCreate)
				log.Println("ContainerCreateQ ============", e.ID, e.String())
				getDetail("ContainerCreateQ", e.ID)
			}
		}
	}
}

type ContainerCreate_RuntimeQ struct {
	*EventQ
}

func (q *ContainerCreate_RuntimeQ) Run() error {
	log.Println("线程ID:", unix.Gettid(), goid.Get())
	for {
		select {
		case <-q.EventQ.ctx.Done():
			return nil
		default:
			item, isShutdown := q.Get()
			if isShutdown {
				log.Println("队列已关闭")
				return nil
			}
			{
				defer q.Done(item)
				e := item.(*eventtypes.ContainerCreate_Runtime)
				log.Println("ContainerCreate_RuntimeQ ============", e.Name, e.String())
			}
		}
	}
}

type ContainerDeleteQ struct {
	*EventQ
}

func (q *ContainerDeleteQ) Run() error {
	log.Println("线程ID:", unix.Gettid(), goid.Get())
	for {
		select {
		case <-q.EventQ.ctx.Done():
			return nil
		default:
			item, isShutdown := q.Get()
			if isShutdown {
				log.Println("队列已关闭")
				return nil
			}
			{
				defer q.Done(item)
				e := item.(*eventtypes.ContainerDelete)
				log.Println("ContainerDeleteQ ============", e.ID, e.String())
			}
		}
	}
}

type ContainerUpdateQ struct {
	*EventQ
}

func (q *ContainerUpdateQ) Run() error {
	log.Println("线程ID:", unix.Gettid(), goid.Get())
	for {
		select {
		case <-q.EventQ.ctx.Done():
			return nil
		default:
			item, isShutdown := q.Get()
			if isShutdown {
				log.Println("队列已关闭")
				return nil
			}
			{
				defer q.Done(item)
				e := item.(*eventtypes.ContainerUpdate)
				log.Println("ContainerUpdateQ ============", e.ID, e.String(), e.Labels)
				getDetail("ContainerUpdateQ", e.ID)
			}
		}
	}
}

func getDetail(task, cid string) {
	client, ctx, cancel, err := NewK8sContainerClient(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer cancel()
	c, err := client.LoadContainer(ctx, cid)
	if err != nil {
		log.Println(fmt.Errorf("get container %s err: %v", cid, err))
		return
	}
	info, err := containerinspector.Inspect(ctx, c)
	if err != nil {
		log.Println(err)
		return
	}

	if isSandbox(info) {
		log.Println("sandbox 不处理", cid)
		return
	}
	log.Println(task, cid, info.Labels)
}

func isSandbox(c *native.Container) bool {
	if t, ok := c.Labels[CONTAINER_TYPE]; ok {
		return t == CONTAINER_TYPE_SANDBOX
	}
	return false
}

func Subscribe(ctx context.Context) error {
	client, ctx, cancel, err := clientutil.NewClient(ctx, "k8s.io", "/run/containerd/containerd.sock")
	if err != nil {
		return err
	}
	defer cancel()

	var eventQueue = NewWorkQueue(ctx)
	eventQueue.Run()

	eventCh, errCh := client.Subscribe(ctx, "")

	go func() {
		for {
			select {
			case e := <-eventCh:
				if strings.HasPrefix(e.Topic, "/tasks/") {
					_, err := convertTaskEvent(e, eventQueue)
					if err != nil {
						//log.Println(cid, err)
						continue
					}
				}
				if strings.HasPrefix(e.Topic, "/containers/") {
					_, err := convertContainerEvent(e, eventQueue)
					if err != nil {
						//log.Println(cid, err)
						continue
					}
				}
			}
		}
	}()

	// 收到错误（一般是 EOF）打印后退出，等待 pod 重新拉起容器
	return <-errCh
}

func convertTaskEvent(msg *events.Envelope, eventQueue *WorkQueue) (string, error) {
	id := ""
	evt, err := typeurl.UnmarshalAny(msg.Event)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshalany: %w", err)
	}

	switch e := evt.(type) {
	case *eventtypes.TaskExecAdded:
		id = e.ContainerID
		eventQueue.TaskExecAdded.Add(e)
	case *eventtypes.TaskExecStarted:
		id = e.ContainerID
		eventQueue.TaskExecStarted.Add(e)
	case *eventtypes.TaskCreate:
		id = e.ContainerID
		eventQueue.TaskCreate.Add(e)
	case *eventtypes.TaskDelete:
		id = e.ContainerID
		eventQueue.TaskDelete.Add(e)
	case *eventtypes.TaskExit:
		id = e.ContainerID
		eventQueue.TaskExit.Add(e)
	case *eventtypes.TaskPaused:
		id = e.ContainerID
		eventQueue.TaskPaused.Add(e)
	case *eventtypes.TaskResumed:
		id = e.ContainerID
		eventQueue.TaskResumed.Add(e)
	case *eventtypes.TaskOOM:
		id = e.ContainerID
		eventQueue.TaskOOM.Add(e)
	default:
		return "", errors.New("unsupported event")
	}
	return id, nil
}

func convertContainerEvent(msg *events.Envelope, eventQueue *WorkQueue) (string, error) {
	id := ""

	evt, err := typeurl.UnmarshalAny(msg.Event)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshalany: %w", err)
	}

	switch e := evt.(type) {
	case *eventtypes.ContainerCreate:
		id = e.ID
		eventQueue.ContainerCreate.Add(e)
	case *eventtypes.ContainerCreate_Runtime:
		id = e.Name
		eventQueue.ContainerCreate_Runtime.Add(e)
	case *eventtypes.ContainerDelete:
		id = e.ID
		eventQueue.ContainerDelete.Add(e)
	case *eventtypes.ContainerUpdate:
		id = e.ID
		eventQueue.ContainerUpdate.Add(e)
	default:
		return "", errors.New("unsupported event")
	}
	return id, nil
}
