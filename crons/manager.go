package crons

import (
	// "context"
	"fmt"
	"time"

	"github.com/d2jvkpn/go-web/pkg/wrap"
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Manager struct {
	cron   *cron.Cron
	logger *zap.Logger
	tasks  []*Task
}

func NewManager(logger *zap.Logger) *Manager {
	return &Manager{
		cron:   cron.New(),
		logger: logger,
		tasks:  make([]*Task, 0, 8),
	}
}

func (m *Manager) Entry(id cron.EntryID) cron.Entry {
	return m.cron.Entry(id)
}

func (m *Manager) intoRun(item *Task) func() {
	return func() {
		entry := m.Entry(item.Id)

		m.logger.Info(
			"running task",
			zap.Int("entryID", int(item.Id)),
			zap.String("name", item.Name),
			zap.String("prev", entry.Prev.Format(time.RFC3339)),
			zap.String("next", entry.Next.Format(time.RFC3339)),
		)

		item.Run()
	}
}

func (m *Manager) AddTask(item Task) (err error) {
	if err = item.Compile(); err != nil {
		return err
	}

	if item.Id, err = m.cron.AddFunc(item.CronExpr, m.intoRun(&item)); err != nil {
		return err
	}

	item.Status = Created
	item.WithLogger(m.logger.Named(fmt.Sprintf("EntryId::%d", item.Id)))
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

	if id <= 0 {
		return fmt.Errorf("invalid id(entryID)")
	}

	eId = cron.EntryID(id)
	for i, v := range m.tasks {
		if v.Id == eId {
			idx, task = i, v
			break
		}
	}
	if task == nil {
		// m.logger.Warn("task not found", zap.Int("id", id))
		return fmt.Errorf("task not found")
	}

	m.cron.Remove(eId)
	task.Remove(by, reason)
	m.tasks = append(m.tasks[:idx], m.tasks[idx+1:]...)
	m.logger.Warn("removing task", zap.Int("entryID", id))

	return nil
}

func (m *Manager) IntoTaskX1(item *Task) *TaskX1 {
	entry := m.Entry(item.Id)
	return &TaskX1{Task: item, Prev: entry.Prev, Next: entry.Next}
}

func (m *Manager) FindTask(id int) (task *TaskX1, err error) {
	var eId cron.EntryID

	if id <= 0 {
		return nil, fmt.Errorf("invalid id(entryID)")
	}

	eId = cron.EntryID(id)
	for _, v := range m.tasks {
		if v.Id == eId {
			return m.IntoTaskX1(v), nil
		}
	}

	return nil, fmt.Errorf("task not found")
}

func (m *Manager) FindAllTasks() (tasks []*TaskX1) {
	tasks = make([]*TaskX1, 0, len(m.tasks))

	for i := range m.tasks {
		tasks = append(tasks, m.IntoTaskX1(m.tasks[i]))
	}

	return tasks
}

func (m *Manager) Start() {
	for i := range m.tasks {
		if m.tasks[i].StartImmediately {
			go m.intoRun(m.tasks[i])()
		}
	}

	m.logger.Info("start cron", zap.Int("numberOfTasks", len(m.tasks)))
	m.cron.Start()
}

func (m *Manager) Shutdown() {
	m.cron.Stop()

	for _, v := range m.tasks {
		m.logger.Info("removing task", zap.Int("entryID", int(v.Id)))
		_ = v.Remove("manager", "shutdown")
	}
	m.logger.Info("shutdown cron", zap.Int("numberOfTasks", len(m.tasks)))
}
