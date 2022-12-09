package models

import (
	// "context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/d2jvkpn/go-web/pkg/misc"
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

func (s *Status) MarshalJSON() ([]byte, error) {
	switch *s {
	case Created, Running, Failed, Cancelled, Removed, Done:
		return []byte(string(*s)), nil
	default:
		return nil, fmt.Errorf("unknow status: %v", s)
	}
}

func (s *Status) UnmarshalJSON(data []byte) error {
	switch Status(string(data)) {
	case Created:
		*s = Created
	case Running:
		*s = Running
	case Failed:
		*s = Failed
	case Cancelled:
		*s = Cancelled
	case Removed:
		*s = Removed
	case Done:
		*s = Done
	default:
		return fmt.Errorf("invalid status")
	}

	return nil
}

type Task struct {
	Name    string   `mapstructure:"name" json:"name,omitempty"` //*
	Path    string   `mapstructure:"path" json:"path,omitempty"` //*
	Args    []string `mapstructure:"args" json:"args,omitempty"`
	WorkDir string   `mapstructure:"work_dir" json:"workDir,omitempty"`
	Cron    Cron     `mapstructure:"cron" json:"cron,omitempty"` //*

	StartImmediately bool `mapstructure:"start_immediately" json:"startImmediately,omitempty"`
	MaxRetries       uint `mapstructure:"max_retries" json:"maxRetries,omitempty"`

	Id        cron.EntryID `json:"id,omitempty"`
	StartAt   time.Time    `json:"startAt,omitempty"`
	UpdatedAt time.Time    `json:"updatedAt,omitempty"`
	CronExpr  string       `json:"cronExpr,omitempty"`
	Pid       int          `json:"pid,omitempty"`
	Status    Status       `json:"status,omitempty"`
	Error     string       `json:"error,omitempty"`

	cmd    *exec.Cmd
	mutex  *sync.RWMutex
	logger *zap.Logger
}

type Cron struct {
	Minute   string `mapstructure:"minute" json:"minute,omitempty"`
	Hour     string `mapstructure:"hour" json:"hour,omitempty"`
	MonthDay string `mapstructure:"month_day" json:"monthDay,omitempty"`
	Month    string `mapstructure:"month" json:"month,omitempty"`
	WeekDay  string `mapstructure:"week_day" json:"weekDay,omitempty"`
}

func (item *Task) Compile() (err error) {
	if item.Name == "" {
		return fmt.Errorf("name is empty")
	}

	item.CronExpr = item.Cron.cronExpr()

	if _, err = cron.ParseStandard(item.CronExpr); err != nil {
		return err
	}

	item.mutex = new(sync.RWMutex)

	// item.cmd = exec.CommandContext(ctx, item.Path, item.Args...)
	item.cmd = exec.Command(item.Path, item.Args...)
	item.cmd.Dir = item.WorkDir
	return nil
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

func (item *Task) UpdateStatus(status Status, err error) (ok bool) {
	item.mutex.Lock()
	ok = item.updateStatus(status, err)
	item.mutex.Unlock()
	return ok
}

func (item *Task) updateStatus(status Status, err error) (ok bool) {
	ok = false

	status2 := map[Status][]Status{
		Created:   {Running},
		Running:   {Failed, Cancelled, Removed, Done},
		Failed:    {Running, Removed},
		Cancelled: {Running, Removed},
		Removed:   {},
		Done:      {Running, Removed},
	}

	ok = misc.VectorIndex[Status](status2[item.Status], status) > -1
	if !ok {
		item.logger.Warn(
			"can't update status",
			zap.String("from", string(item.Status)),
			zap.String("to", string(status)),
		)
		return
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

	return ok
}

func (item *Task) GetStatus() (s Status) {
	item.mutex.RLock()
	defer item.mutex.RUnlock()

	s = item.Status
	return s
}

func (item *Task) RLock() {
	item.mutex.RLock()
}

func (item *Task) RUnlock() {
	item.mutex.RUnlock()
}

func (item *Task) Run() {
	if item.StartAt.IsZero() {
		now := time.Now()
		item.StartAt, item.UpdatedAt = now, now
	}

	go func() {
		var (
			pid    int
			err    error
			status Status
		)

		if status = item.GetStatus(); status == Removed {
			return
		}

		if status == Running {
			_, err = item.kill()
			item.UpdateStatus(Cancelled, err)
		}

		for i := 0; i < int(item.MaxRetries)+1; i++ {
			if i > 0 {
				time.Sleep(3 * time.Second)
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
			item.UpdateStatus(Done, nil)
		}
	}()
}

func (item *Task) Remove(by, reason string) bool {
	if item.GetStatus() != Running {
		return false
	}

	item.mutex.Lock()
	_, _ = item.kill()
	item.updateStatus(Removed, fmt.Errorf("by: %q, reason: %q", by, reason))
	item.mutex.Unlock()
	return true
}

func (item *Task) kill() (ok bool, err error) {
	defer item.logger.Warn("kill process", zap.Bool("ok", ok), zap.Any("error", err))
	// TODO send a term signal to process
	if item.GetStatus() == Running {
		ok = true
		err = item.cmd.Process.Kill()
		return
	}

	ok = false
	return
}
