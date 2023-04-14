package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/ftlynx/httpbash/internal/global"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

const (
	ProcessStatusFail    string = "fail"
	ProcessStatusSuccess string = "success"
	ProcessStatusRunning string = "running"
)

var ErrTaskIsRunning = errors.New("task is running")

var taskLock sync.Map

type Command struct {
	Name        string            `json:"name"`
	Args        []string          `json:"args"`
	Timout      time.Duration     `json:"timout"`
	ConfigFile  CommandConfigFile `json:"config_file"`
	CreatedUser string            `json:"created_user"`
}

type task struct {
	DataDir string  `json:"data_dir"`
	TaskId  string  `json:"task_id"`
	JobId   string  `json:"job_id"`
	Command Command `json:"command"`
	file    *os.File
}

func (t *task) TaskIsRunning() bool {
	_, ok := taskLock.Load(t.TaskId)
	if ok {
		return true
	}
	return false
}

func (t *task) SetTaskLock(isLock bool) {
	if isLock {
		taskLock.Store(t.TaskId, true)
	} else {
		taskLock.Delete(t.TaskId)
	}
}

func (t *task) SetCommand(command Command) TaskDao {
	t.Command = command
	return t
}

func (t *task) StoreJobConfigFile() error {
	return t.Command.ConfigFile.StoreFile("config", t.DataDir)
}

func (t *task) GetJobConfigFilePath() string {
	return fmt.Sprintf("%s/config", t.DataDir)
}

func (t *task) GetJobLogFilePath() string {
	return fmt.Sprintf("%s/log", t.DataDir)
}

func (t *task) GetJobStatusFilePath() string {
	return fmt.Sprintf("%s/status", t.DataDir)
}

func (t *task) StoreJobStatus(status string) error {
	return os.WriteFile(t.GetJobStatusFilePath(), []byte(status), 0666)
}

func (t *task) GetJobStatus() ([]byte, error) {
	return os.ReadFile(t.GetJobStatusFilePath())
}
func (t *task) ReadJobFullLog() ([]byte, error) {
	return os.ReadFile(t.GetJobLogFilePath())
}

func (t *task) SetFile(file *os.File) TaskDao {
	t.file = file
	return t
}
func (t *task) ReadJobNextLog() ([]byte, error) {
	var err error
	_, err = t.file.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 1024) // 缓冲区
	count, err := t.file.Read(buf)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return buf[:count], nil
}

func (t *task) Exec() error {
	log.SetOutput(os.Stdout)
	log.SetPrefix(fmt.Sprintf("%s %s", t.TaskId, t.JobId))

	if err := os.MkdirAll(t.DataDir, 0755); err != nil {
		return err
	}

	if t.TaskIsRunning() {
		log.Printf("task is running. can not exec")
		return ErrTaskIsRunning
	}

	if err := t.StoreJobConfigFile(); err != nil {
		return err
	}

	go func() {
		// 解锁任务
		defer func() {
			t.SetTaskLock(false)
		}()

		// 锁定任务
		t.SetTaskLock(true)

		jobStatus := ProcessStatusFail
		defer func() {
			if err := t.StoreJobStatus(jobStatus); err != nil {
				log.Printf("store job status fail %s", err.Error())
				return
			}
		}()

		opFile, err := os.OpenFile(t.GetJobLogFilePath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Printf("openfile %s fail, %s", t.GetJobLogFilePath(), err.Error())
			return
		}
		defer opFile.Close()

		ctx, cancel := context.WithTimeout(context.Background(), t.Command.Timout)
		defer cancel()

		cmd := exec.CommandContext(ctx, t.Command.Name, t.Command.Args...)
		log.Printf("[%s] exec begin\n", cmd.String())

		if _, err := opFile.WriteString(fmt.Sprintf("author: %s\ncommand: %s\n\n", t.Command.CreatedUser, cmd.String())); err != nil {
			log.Printf("file write command fail %s", err.Error())
			return
		}

		cmd.Stdout = opFile
		cmd.Stderr = opFile

		if err := cmd.Run(); err != nil {
			log.Printf("[%s] cmd run err: %s\n", cmd.String(), err.Error())
			_, _ = opFile.WriteString(fmt.Sprintf("\n\nprocess run err: %s", err.Error()))
			return
		}

		log.Printf("[%s] exec success\n", cmd.String())
		jobStatus = ProcessStatusSuccess
	}()

	return nil
}

type TaskDao interface {
	GetJobConfigFilePath() string
	GetJobLogFilePath() string
	ReadJobFullLog() ([]byte, error)
	ReadJobNextLog() ([]byte, error)
	GetJobStatus() ([]byte, error)
	TaskIsRunning() bool
	SetCommand(command Command) TaskDao
	SetFile(file *os.File) TaskDao
	Exec() error
}

func NewTask(taskId string, jobId string) TaskDao {
	dataDir := fmt.Sprintf("%s/%s/%s", global.Conf.App.DataDir, taskId, jobId)
	return &task{
		DataDir: dataDir,
		TaskId:  taskId,
		JobId:   jobId,
	}
}
