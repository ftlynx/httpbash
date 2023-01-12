package main

import "gorm.io/gorm"

type TaskLog struct {
	DB    *gorm.DB
	UUID  string
	count int64 // 日志字符数
}

func (t *TaskLog) GetNewLog() (uint8, string, error) {
	result := Task{}
	if err := t.DB.Where("uuid=?", t.UUID).First(&result).Error; err != nil {
		return ProcessStatusRunning, "", err
	}
	if t.count == 0 {
		t.count = int64(len(result.ProcessLog))
		return result.ProcessStatus, result.ProcessLog, nil
	}
	newLog := result.ProcessLog[t.count:]
	t.count = int64(len(result.ProcessLog))
	return result.ProcessStatus, newLog, nil
}
