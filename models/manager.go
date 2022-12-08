package models

import (
	// "context"
	"fmt"
	"os"

	"github.com/d2jvkpn/go-web/pkg/wrap"
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
)

type Manager struct {
	cron  *cron.Cron
	Pid   int
	tasks []*Task
}

func NewManager() *Manager {
	return &Manager{
		cron:  cron.New(),
		Pid:   os.Getpid(),
		tasks: make([]*Task, 0, 8),
	}
}

func (m *Manager) AddTask(item Task) (err error) {
	if err = item.Compile(); err != nil {
		return err
	}

	if item.Id, err = m.cron.AddFunc(item.CronExpr, item.Run); err != nil {
		return err
	}

	m.tasks = append(m.tasks, &item)

	return nil
}

func (m *Manager) LoadTasksFronConfig(p, field string) (num int, err error) {
	var (
		cv    *viper.Viper
		tasks []Task
	)

	if cv, err = wrap.OpenConfig("configs/test.yaml"); err != nil {
		return 0, err
	}

	if err = cv.UnmarshalKey("jobs", &tasks); err != nil {
		return 0, err
	}

	for i := range tasks {
		if err = m.AddTask(tasks[i]); err != nil {
			return num, err
		}
		num++
	}

	return num, nil
}

func (m *Manager) RemoveTask(id int) (err error) {
	var (
		idx  int
		eId  cron.EntryID
		task *Task
	)

	eId = cron.EntryID(id)
	for i, v := range m.tasks {
		if v.Id == eId {
			idx, task = i, v
			break
		}
	}
	if task == nil {
		return fmt.Errorf("task not found")
	}

	m.cron.Remove(eId)
	task.Remove()
	m.tasks = append(m.tasks[:idx], m.tasks[idx+1:]...)
	return nil
}

func (m *Manager) CloneTasks() (tasks []Task) {
	tasks = make([]Task, 0, len(m.tasks))

	for i := range m.tasks {
		v := m.tasks[i]
		if v.GetStatus() == Removed {
			continue
		}

		v.RLock()
		t := *v
		t.cmd, t.mutex = nil, nil
		v.RUnlock()
		tasks = append(tasks, t)
	}

	return tasks
}

func (m *Manager) Start() {
	for i := range m.tasks {
		if m.tasks[i].StartImmediately {
			m.tasks[i].Run()
		}
	}
	m.cron.Start()
}

func (m *Manager) Shutdown() {
	m.cron.Stop()

	for _, v := range m.tasks {
		v.Remove()
	}
}
