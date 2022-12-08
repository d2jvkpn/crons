package models

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
	Restart          int  `mapstructure:"restart" json:"restart,omitempty"`
	// AllowOverlap     bool `mapstructure:"allow_overlap" json:"allowOverlap,omitempty"`
	// ?? Times int             // => max run times
	// ?? Timeout time.Duration // => timeout for each run

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
	from := item.Status

	switch status {
	case Created, Running:
		ok = true
	case Failed, Cancelled, Removed:
		if item.Status == Running {
			ok = true
		}
	case Done:
		switch item.Status {
		case Running, Cancelled, Removed:
			ok = true
		}
	}

	if ok {
		item.Status, item.UpdatedAt = status, time.Now()
		if err != nil {
			item.Error = err.Error()
		}
	}

	item.logger.Warn(
		"update status",
		zap.String("from", string(from)),
		zap.String("from", string(status)),
		zap.String("error", item.Error),
	)

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
			err = item.cmd.Process.Kill()
			item.UpdateStatus(Cancelled, err)
		}

		err = item.cmd.Start()
		if err != nil {
			item.logger.Warn("start", zap.String("error", err.Error()))
			item.UpdateStatus(Failed, err)
			return
		}
		item.UpdateStatus(Running, nil)
		if pid = 0; item.cmd.Process != nil {
			pid = item.cmd.Process.Pid
			item.Pid = pid
		}
		item.logger.Warn("start", zap.Int("pid", pid))

		if err = item.cmd.Wait(); err != nil {
			item.UpdateStatus(Failed, err)
		} else {
			item.UpdateStatus(Done, nil)
		}
	}()
}

func (item *Task) Remove() bool {
	if item.GetStatus() != Running {
		return false
	}

	item.mutex.Lock()
	item.cmd.Process.Kill() // TODO send a term signal to process
	item.updateStatus(Removed, nil)
	item.mutex.Unlock()
	return true
}
