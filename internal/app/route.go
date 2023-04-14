package app

import (
	"bytes"
	"embed"
	"fmt"
	"github.com/ftlynx/httpbash/internal/global"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"
)

// 注意 仅能嵌入当前目录及其子目录，无法嵌入上层目录。同时也不支持软链接
//
//go:embed dist/*
var staticFiles embed.FS

type CommandExecBody struct {
	TaskId        string            `json:"task_id" binding:"required"` // 任务ID，标识是否是同一个任务
	Cmd           string            `json:"cmd" binding:"required"`
	ConfigFile    CommandConfigFile `json:"config_file"`
	TimeoutMinute int64             `json:"timeout_minute" binding:"required"`
	CreatedUser   string            `json:"created_user"`
}
type Response struct {
	Success      bool        `json:"success"`
	ErrorMessage string      `json:"error_message,omitempty"`
	Data         interface{} `json:"data,omitempty"`
}

func NewFail(errorMessage string) Response {
	return Response{
		Success:      false,
		ErrorMessage: errorMessage,
	}
}

func NewOk(data interface{}) Response {
	return Response{
		Success: true,
		Data:    data,
	}
}

var upgrader = websocket.Upgrader{
	HandshakeTimeout: 2 * time.Second,
	ReadBufferSize:   1024,
	WriteBufferSize:  1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {},
}

func MyRouter() error {
	r := gin.Default()
	noAuth := r.Group("/v0")
	staticFS, err := fs.Sub(staticFiles, "dist")
	if err != nil {
		return err
	}
	noAuth.StaticFS("/console", http.FS(staticFS))
	auth := r.Group("/v1")
	auth.Use(func(c *gin.Context) {
		apiAuth := c.GetHeader("x-api-auth")
		if apiAuth != global.Conf.App.Auth {
			c.JSON(http.StatusForbidden, NewFail("auth fail"))
			c.Abort()
			return
		}
	})

	auth.POST("/command", func(c *gin.Context) {
		body := CommandExecBody{}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, NewFail(err.Error()))
			return
		}
		// 超时时间默认 60分钟
		if body.TimeoutMinute == 0 {
			body.TimeoutMinute = 60
		}

		jobId := fmt.Sprintf("%d", time.Now().UnixMilli())
		task := NewTask(body.TaskId, jobId)

		// 替换配置文件变量
		tmpl, err := template.New("config").Parse(body.Cmd)
		if err != nil {
			c.JSON(http.StatusBadRequest, NewFail(err.Error()))
			return
		}
		var cmdBytes bytes.Buffer
		tp := map[string]string{
			"config": task.GetJobConfigFilePath(),
		}
		if err := tmpl.Execute(&cmdBytes, tp); err != nil {
			c.JSON(http.StatusBadRequest, NewFail(err.Error()))
			return
		}

		cmdSlice := strings.Fields(strings.TrimSpace(cmdBytes.String())) // 先去除前后的空格，然后在按空格分隔
		if !InSlice(cmdSlice[0], global.Conf.Command.Whitelist) {
			c.JSON(http.StatusForbidden, NewFail(fmt.Sprintf("command %s non-existent command.whitelist", cmdSlice[0])))
			return
		}

		command := Command{
			Name:        cmdSlice[0],
			Args:        cmdSlice[1:],
			ConfigFile:  body.ConfigFile,
			Timout:      time.Duration(body.TimeoutMinute) * time.Minute,
			CreatedUser: body.CreatedUser,
		}
		task.SetCommand(command)

		if err := task.Exec(); err != nil {
			c.JSON(http.StatusForbidden, NewFail(err.Error()))
			return
		}
		wsPrefix := ""
		if strings.HasPrefix(global.Conf.App.DisplayUrl, "https://") {
			wsPrefix = strings.Replace(global.Conf.App.DisplayUrl, "https://", "wss://", 1)
		}
		if strings.HasPrefix(global.Conf.App.DisplayUrl, "http://") {
			wsPrefix = strings.Replace(global.Conf.App.DisplayUrl, "http://", "ws://", 1)
		}

		result := make(map[string]string)
		result["http_api_endpoint"] = fmt.Sprintf("%s/v0/command/log?task_id=%s&job_id=%s", global.Conf.App.DisplayUrl, body.TaskId, jobId)
		result["ws_api_endpoint"] = fmt.Sprintf("%s/v0/command/log/ws?task_id=%s&job_id=%s", wsPrefix, body.TaskId, jobId)
		result["ws_html_endpoint"] = fmt.Sprintf("%s/v0/console/?task_id=%s&job_id=%s", global.Conf.App.DisplayUrl, body.TaskId, jobId)
		c.JSON(http.StatusOK, NewOk(result))
		return
	})
	noAuth.GET("/command/log", func(c *gin.Context) {
		taskId := c.Query("task_id")
		if taskId == "" {
			c.JSON(http.StatusBadRequest, NewFail("task_id require"))
			return
		}
		jobId := c.Query("job_id")
		if jobId == "" {
			c.JSON(http.StatusBadRequest, NewFail("job_id require"))
			return
		}

		task := NewTask(taskId, jobId)
		data, err := task.ReadJobFullLog()
		if err != nil {
			c.JSON(http.StatusInternalServerError, NewFail(err.Error()))
			return
		}
		jobStatus := ProcessStatusRunning
		if task.TaskIsRunning() == false {
			status, err := task.GetJobStatus()
			if err != nil {
				c.JSON(http.StatusInternalServerError, NewFail("job not exists"))
				return
			}
			jobStatus = string(status)
		}

		c.JSON(http.StatusOK, NewOk(gin.H{
			"content": string(data),
			"status":  jobStatus,
		}))
		return
	})
	noAuth.GET("/command/log/ws", func(c *gin.Context) {
		taskId := c.Query("task_id")
		if taskId == "" {
			c.JSON(http.StatusBadRequest, NewFail("task_id require"))
			return
		}
		jobId := c.Query("job_id")
		if jobId == "" {
			c.JSON(http.StatusBadRequest, NewFail("job_id require"))
			return
		}

		task := NewTask(taskId, jobId)
		file, err := os.Open(task.GetJobLogFilePath())
		if err != nil {
			c.JSON(http.StatusInternalServerError, NewFail(err.Error()))
			return
		}
		defer file.Close()

		task.SetFile(file)

		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.JSON(http.StatusMethodNotAllowed, NewFail(err.Error()))
			return
		}
		defer ws.Close()

		for {
			data, err := task.ReadJobNextLog()
			if err != nil {
				c.JSON(http.StatusInternalServerError, NewFail(err.Error()))
				return
			}
			if err == io.EOF && !task.TaskIsRunning() {
				// 读取完文件，但是任务状态还是运行中，可能产生新的日志
				break
			}
			if len(data) == 0 {
				// 任务未结束，但是没有新日志
				time.Sleep(2 * time.Second)
				continue
			}
			if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
				c.JSON(http.StatusInternalServerError, NewFail(err.Error()))
				return
			}
		}
	})

	return r.Run(global.Conf.App.Listen)
}
