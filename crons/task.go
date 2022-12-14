package crons

import (
	// "context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type Status string

const (
	Created   Status = "created"
	Running   Status = "running"
	Failed    Status = "failed"
	Cancelled Status = "cancelled"
	Removed   Status = "removed"
	Done      Status = "done"
)

//func (s *Status) MarshalJSON() ([]byte, error) {
//	switch *s {
//	case Created, Running, Failed, Cancelled, Removed, Done:
//		return []byte(string(*s)), nil
//	default:
//		return nil, fmt.Errorf("unknow status: %v", s)
//	}
//}

//func (s *Status) UnmarshalJSON(data []byte) error {
//	status := Status(string(data))
//	switch status {
//	case Created, Running, Failed, Cancelled, Done, Removed:
//		*s = status
//	default:
//		return fmt.Errorf("invalid status")
//	}

//	return nil
//}

type Task struct {
	Name    string   `mapstructure:"name" json:"name,omitempty"` //*
	Path    string   `mapstructure:"path" json:"path,omitempty"` //*
	Args    []string `mapstructure:"args" json:"args,omitempty"`
	WorkDir string   `mapstructure:"work_dir" json:"workDir,omitempty"`
	Cron    Cron     `mapstructure:"cron" json:"cron,omitempty"` //*

	StartImmediately bool `mapstructure:"start_immediately" json:"startImmediately,omitempty"`
	MaxRetries       uint `mapstructure:"max_retries" json:"maxRetries,omitempty"`

	Id        cron.EntryID `json:"id,omitempty"`
	CreatedAt time.Time    `json:"createdAt,omitempty"`
	StartAt   time.Time    `json:"startAt,omitempty"`
	UpdatedAt time.Time    `json:"updatedAt,omitempty"`
	CronExpr  string       `json:"cronExpr,omitempty"`
	Pid       int          `json:"pid,omitempty"`
	Status    Status       `json:"status,omitempty"`
	Error     string       `json:"error,omitempty"`

	cmd    *exec.Cmd
	mutex  *sync.RWMutex
	logger *zap.Logger
	ch     chan struct{}
}

type Cron struct {
	Minute   string `mapstructure:"minute" json:"minute,omitempty"`
	Hour     string `mapstructure:"hour" json:"hour,omitempty"`
	MonthDay string `mapstructure:"month_day" json:"monthDay,omitempty"`
	Month    string `mapstructure:"month" json:"month,omitempty"`
	WeekDay  string `mapstructure:"week_day" json:"weekDay,omitempty"`
}

func (item *Task) Clone(clear bool) (task Task) {
	task = *item
	task.cmd, task.mutex = nil, nil
	task.logger, task.ch = nil, nil
	if !clear {
		return
	}
	item.setDefault()
	return
}

func (item *Task) setDefault() {
	var at time.Time
	item.Id = cron.EntryID(0)
	item.StartAt, item.UpdatedAt = at, at
	item.Pid, item.Status, item.Error = 0, Created, ""

	at = time.Now()
	item.CreatedAt = at
}

func (item *Task) Compile() (err error) {
	if item.Name == "" {
		return fmt.Errorf("name is empty")
	}

	item.CronExpr = item.Cron.cronExpr()

	if _, err = cron.ParseStandard(item.CronExpr); err != nil {
		return err
	}

	item.setDefault()
	item.mutex = new(sync.RWMutex)
	item.ch = make(chan struct{}, 1)
	item.setCmd()

	return nil
}

func (item *Task) setCmd() {
	// item.cmd = exec.CommandContext(ctx, item.Path, item.Args...)
	item.cmd = exec.Command(item.Path, item.Args...)
	item.cmd.Dir = item.WorkDir
}

func (item *Task) WithLogger(logger *zap.Logger) *Task {
	item.logger = logger
	return item
}

func (item *Cron) cronExpr() string {
	if v := &item.Minute; *v == "" {
		*v = "*"
	}
	if v := &item.Hour; *v == "" {
		*v = "*"
	}
	if v := &item.MonthDay; *v == "" {
		*v = "*"
	}
	if v := &item.Month; *v == "" {
		*v = "*"
	}
	if v := &item.WeekDay; *v == "" {
		*v = "*"
	}

	// crontab: minute, hour, month_day, month, week_day
	return fmt.Sprintf(
		"%s %s %s %s %s",
		item.Minute, item.Hour, item.MonthDay, item.Month, item.WeekDay,
	)
}

func (item *Task) UpdateStatus(status Status, err error) {
	item.mutex.Lock()
	item.updateStatus(status, err)
	item.mutex.Unlock()
	return
}

func (item *Task) updateStatus(status Status, err error) {
	if status != Running {
		item.Pid = 0
	}
	if err != nil {
		item.Error = err.Error()
	} else {
		item.Error = ""
	}

	fields := []zap.Field{
		zap.String("from", string(item.Status)),
		zap.String("to", string(status)),
		zap.String("error", item.Error),
	}

	item.Status, item.UpdatedAt = status, time.Now()

	switch status {
	case Failed:
		item.logger.Error("update status", fields...)
	case Cancelled, Removed:
		item.logger.Warn("update status", fields...)
	default:
		item.logger.Info("update status", fields...)
	}
}

func (item *Task) GetStatus() (s Status) {
	item.mutex.RLock()
	defer item.mutex.RUnlock()

	s = item.getStatus()
	return s
}

func (item *Task) getStatus() (s Status) {
	s = item.Status
	return s
}

func (item *Task) Run() {
	var (
		pid    int
		err    error
		status Status
		now    time.Time
	)

	if status = item.GetStatus(); status == Removed {
		return
	}

	if status == Running {
		_, err = item.kill()
		item.ch <- struct{}{} // wait for the previous goroutine to exit
		item.UpdateStatus(Cancelled, err)
	} else {
		item.ch <- struct{}{}
	}
	defer func() {
		<-item.ch
	}()

	now = time.Now()
	item.UpdatedAt = now
	item.setCmd()
	for i := 0; i < int(item.MaxRetries)+1; i++ {
		if i > 0 {
			time.Sleep(RetryAfter)
		}

		if err = item.cmd.Start(); err != nil {
			item.UpdateStatus(Failed, err)
		} else {
			break
		}
	}
	if err != nil {
		item.logger.Error("abort task", zap.Uint("retryTimes", item.MaxRetries))
		return
	}

	item.UpdateStatus(Running, nil)
	if pid = 0; item.cmd.Process != nil {
		pid = item.cmd.Process.Pid
		item.Pid = pid
	}
	item.logger.Info("started task", zap.Int("pid", pid))

	if err = item.cmd.Wait(); err != nil {
		item.UpdateStatus(Failed, err)
	} else {
		item.StartAt = now
		item.UpdateStatus(Done, nil)
	}
}

func (item *Task) Remove(by, reason string) bool {
	if item.GetStatus() != Running {
		return false
	}

	item.mutex.Lock()
	_, _ = item.kill()
	item.updateStatus(Removed, fmt.Errorf("by:%s, reason:%s", by, reason))
	item.mutex.Unlock()
	return true
}

func (item *Task) kill() (ok bool, err error) {
	defer func() {
		if err != nil {
			item.logger.Warn("kill process", zap.Bool("ok", ok), zap.Any("error", err.Error()))
		} else {
			item.logger.Warn("kill process", zap.Bool("ok", ok))
		}
	}()
	// TODO send a term signal to process
	if item.getStatus() == Running {
		ok = true
		err = item.cmd.Process.Kill()
		return
	}

	ok = false
	return
}
