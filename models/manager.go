package models

import (
	// "context"
	"fmt"
	"os"

	"github.com/d2jvkpn/go-web/pkg/wrap"
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Manager struct {
	cron   *cron.Cron
	Pid    int
	logger *zap.Logger
	tasks  []*Task
}

func NewManager(logger *zap.Logger) *Manager {
	return &Manager{
		cron:   cron.New(),
		Pid:    os.Getpid(),
		logger: logger,
		tasks:  make([]*Task, 0, 8),
	}
}

func (m *Manager) AddTask(item Task) (err error) {
	if err = item.Compile(); err != nil {
		return err
	}

	if item.Id, err = m.cron.AddFunc(item.CronExpr, item.Run); err != nil {
		return err
	}

	item.Status = Created
	item.logger = m.logger.Named(fmt.Sprintf("task_%d", item.Id))
	m.tasks = append(m.tasks, &item)

	m.logger.Info("add task", zap.Any("task", item))

	return nil
}

func (m *Manager) LoadTasksFronConfig(p, field string) (num int, err error) {
	var (
		cv    *viper.Viper
		tasks []Task
	)

	if cv, err = wrap.OpenConfig(p); err != nil {
		return 0, err
	}

	if err = cv.UnmarshalKey(field, &tasks); err != nil {
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

func (m *Manager) RemoveTask(id int, by, reason string) (err error) {
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
		m.logger.Warn("task not found", zap.Int("id", id))
		return fmt.Errorf("task not found")
	}

	m.cron.Remove(eId)
	task.Remove(by, reason)
	m.tasks = append(m.tasks[:idx], m.tasks[idx+1:]...)
	m.logger.Warn("remove task", zap.Int("id", id))

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

	m.logger.Info("Start Cron", zap.Int("pid", m.Pid), zap.Int("numberOfTasks", len(m.tasks)))
	m.cron.Start()
}

func (m *Manager) Shutdown() {
	m.cron.Stop()

	for _, v := range m.tasks {
		_ = v.Remove("manager", "shutdown")
	}
	m.logger.Info("Shutdown Cron", zap.Int("numberOfTasks", len(m.tasks)))
}
