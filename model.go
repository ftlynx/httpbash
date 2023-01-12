package main

const (
	ProcessStatusUnKnown uint8 = 0
	ProcessStatusFail    uint8 = 1
	ProcessStatusRunning uint8 = 2
	ProcessStatusSuccess uint8 = 3
)

var ConfigFileRootDir string

type Task struct {
	UUID          string `gorm:"primarykey" json:"uuid"`
	CmdString     string `gorm:"type:varchar(256)" json:"cmd_string"`
	ProcessLog    string `gorm:"type:longtext" json:"process_log"`
	ProcessStatus uint8  `gorm:"type:tinyint" json:"process_status"`
	CreatedUser   string `gorm:"type:varchar(128);not null" json:"created_user"`
	CreatedAt     int64  `gorm:"autoCreateTime:milli" json:"created_at"`
	UpdatedAt     int64  `gorm:"autoUpdateTime:milli" json:"updated_at"`
}

func (t *Task) TableName() string {
	return "httpbash_task"
}
