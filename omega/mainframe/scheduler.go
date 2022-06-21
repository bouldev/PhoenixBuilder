package mainframe

import (
	"phoenixbuilder/omega/defines"
	"sync"
)

const (
	UrgentTask = uint64(1) << iota
	NormalTask
	BackGroundTask
	Paused
	Executing
	Done
)

type taskNode struct {
	next          *taskNode
	flag          uint64
	task          defines.BotTask
	pauseAbleTask defines.BotTaskPauseAble
}

type OmegaBotTaskScheduler struct {
	mu             sync.Mutex
	taskLink       *taskNode
	urgentTaskHead *taskNode
	normalTaskHead *taskNode
}

func NewOmegaBotTaskScheduler() *OmegaBotTaskScheduler {
	o := &OmegaBotTaskScheduler{
		mu: sync.Mutex{},
	}
	return o
}

func (o *OmegaBotTaskScheduler) CommitBackgroundTask(task defines.BotTaskPauseAble) (reject, pending bool) {
	if o.taskLink != nil {
		return true, false
	}
	return false, o.CommitNormalTask(task)
}

func (o *OmegaBotTaskScheduler) CommitNormalTask(task defines.BotTaskPauseAble) (pending bool) {
	ntn := &taskNode{
		pauseAbleTask: task,
		flag:          NormalTask,
	}
	o.mu.Lock()
	if o.taskLink != nil {
		tn := o.taskLink
		for {
			if tn.next == nil {
				break
			}
			tn = tn.next
		}
		tn.next = ntn
		if o.normalTaskHead == nil {
			o.normalTaskHead = ntn
		}
		o.mu.Unlock()
		return true
	} else {
		o.taskLink = ntn
		o.normalTaskHead = ntn
		o.mu.Unlock()
		go o.scheduleNext()
		return false
	}
}

func (o *OmegaBotTaskScheduler) scheduleNext() {
	o.mu.Lock()
	if o.taskLink == nil {
		o.mu.Unlock()
		return
	}
	if (o.taskLink.flag & UrgentTask) != 0 {
		t := o.taskLink
		t.flag &= Executing
		go func() {
			t.task.Activate()
			t.flag = t.flag & (^Executing)
			t.flag = t.flag | Done
			o.mu.Lock()
			o.taskLink = t.next
			o.urgentTaskHead = t.next
			if o.urgentTaskHead == o.normalTaskHead {
				o.urgentTaskHead = nil
			}
			o.mu.Unlock()
			o.scheduleNext()
		}()
		o.mu.Unlock()
		return
	} else {
		if (o.taskLink.flag & Paused) == 0 {
			t := o.taskLink
			t.flag &= Executing
			go func() {
				t.pauseAbleTask.Activate()
				t.flag = t.flag & (^Executing)
				t.flag |= Done
				o.mu.Lock()
				var lastLinkNode *taskNode
				currentLinkNode := o.taskLink
				o.taskLink = nil
				o.urgentTaskHead = nil
				o.normalTaskHead = nil
				for {
					if currentLinkNode.next == nil {
						break
					}
					if (currentLinkNode.flag & Done) == 0 {
						if lastLinkNode == nil {
							lastLinkNode = currentLinkNode
						} else {
							lastLinkNode.next = currentLinkNode
						}
						if o.taskLink == nil {
							o.taskLink = currentLinkNode
						}
						if o.urgentTaskHead == nil && (currentLinkNode.flag&UrgentTask) != 0 {
							o.urgentTaskHead = currentLinkNode
						}
						if o.normalTaskHead == nil && (currentLinkNode.flag&NormalTask) != 0 {
							o.normalTaskHead = currentLinkNode
						}
					}
					currentLinkNode = currentLinkNode.next
				}
				o.mu.Unlock()
				o.scheduleNext()
			}()
			o.mu.Unlock()
		} else {
			t := o.taskLink
			t.flag = t.flag & (^Paused)
			t.flag = t.flag & Executing
			go func() {
				t.pauseAbleTask.Resume()
			}()
			o.mu.Unlock()
		}
	}
}

func (o *OmegaBotTaskScheduler) CommitUrgentTask(task defines.BotTask) (pending bool) {
	ntn := &taskNode{
		task: task,
		flag: UrgentTask,
	}
	o.mu.Lock()
	if o.normalTaskHead != nil {
		tn := o.normalTaskHead
		if tn.flag&Executing != 0 {
			tn.pauseAbleTask.Pause()
			tn.flag = tn.flag | Paused
			tn.flag = tn.flag & (^Executing)
		}
	}
	if o.urgentTaskHead == nil {
		ntn.next = o.normalTaskHead
		o.taskLink = ntn
		o.urgentTaskHead = ntn
		o.mu.Unlock()
		go o.scheduleNext()
		return false
	} else {
		tn := o.urgentTaskHead
		for {
			if tn.next == o.normalTaskHead {
				break
			}
			tn = tn.next
		}
		ntn.next = tn.next
		tn.next = ntn
		o.mu.Unlock()
		return true
	}
}
