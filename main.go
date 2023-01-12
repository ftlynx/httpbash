package main

import (
	"fmt"
	"github.com/ftlynx/httpbash/version"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type CommandExecBody struct {
	UUID          string            `json:"uuid"` // 用户可自己设置uuid
	Cmd           string            `json:"cmd" binding:"required"`
	ConfigFile    CommandConfigFile `json:"config_file"`
	TimeoutMinute int64             `json:"timeout_minute" binding:"required"`
	CreatedUser   string            `json:"created_user"`
}
type CommandExecResponse struct {
	ExecUUID string `json:"exec_uuid"`
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

func main() {
	rootDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}
	ConfigFileRootDir = rootDir + "/run/config/"

	pflag.BoolP("version", "v", false, "show version")
	pflag.Parse()
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		panic(err)
	}
	if viper.GetBool("version") {
		fmt.Printf(version.FullVersion())
		return
	}

	viper.SetConfigFile("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	commandWhitelist := viper.GetStringSlice("command.whitelist")
	mysqlDatasource := viper.GetString("mysql.datasource")
	appListen := viper.GetString("app.listen")
	if appListen == "" {
		appListen = ":8080"
	}

	appDebug := viper.GetBool("app.debug")
	appAuth := viper.GetString("app.auth")
	gormLogLevel := logger.Silent
	gin.SetMode(gin.ReleaseMode)
	if appDebug {
		gin.SetMode(gin.DebugMode)
		gormLogLevel = logger.Warn
	}
	db, err := gorm.Open(mysql.Open(mysqlDatasource), &gorm.Config{
		Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
			SlowThreshold:             2 * time.Second,
			LogLevel:                  gormLogLevel,
			IgnoreRecordNotFoundError: false,
			Colorful:                  false,
		}),
	})
	if err != nil {
		panic(err)
	}
	err = db.AutoMigrate(&Task{})
	if err != nil {
		panic(err)
	}

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		apiAuth := c.GetHeader("x-api-auth")
		if apiAuth != appAuth {
			c.JSON(http.StatusForbidden, gin.H{"error_message": "not auth"})
			c.Abort()
			return
		}
	})
	r.POST("/command", func(c *gin.Context) {
		body := CommandExecBody{}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error_message": err.Error()})
			return
		}
		// 超时时间默认 60分钟
		if body.TimeoutMinute == 0 {
			body.TimeoutMinute = 60
		}

		// 优先使用用户设置的uuid，避免解析返回内容获取uuid
		if body.UUID == "" {
			body.UUID = uuid.New().String()
		}

		// 配置文件写入本地
		if err := body.ConfigFile.StoreFile(body.UUID, ConfigFileRootDir); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error_message": err.Error()})
			return
		}

		cmdSlice := strings.Fields(strings.TrimSpace(body.Cmd)) // 先去除前后的空格，然后在按空格分隔
		if !InSlice(cmdSlice[0], commandWhitelist) {
			c.JSON(http.StatusForbidden, gin.H{"error_message": fmt.Sprintf("command %s non-existent command.whitelist", cmdSlice[0])})
			return
		}
		myCmd := Command{
			Name:        cmdSlice[0],
			Args:        cmdSlice[1:],
			Timout:      time.Duration(body.TimeoutMinute) * time.Minute,
			UUID:        body.UUID,
			CreatedUser: body.CreatedUser,
			DB:          db,
		}
		go myCmd.Exec()

		c.JSON(http.StatusOK, CommandExecResponse{
			ExecUUID: body.UUID,
		})
		return
	})
	r.GET("/command", func(c *gin.Context) {
		execUUID := c.Query("uuid")
		if execUUID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error_message": "query parameter uuid require"})
			return
		}
		data := Task{}
		if err := db.Where("uuid=?", execUUID).First(&data).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error_message": fmt.Sprintf("get task by uuid(%s) fail", execUUID)})
			return
		}
		if data.ProcessStatus == ProcessStatusRunning {
			c.JSON(http.StatusPartialContent, data)
		} else {
			c.JSON(http.StatusOK, data)
		}
		return
	})
	r.GET("/command/ws", func(c *gin.Context) {
		execUUID := c.Query("uuid")
		if execUUID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error_message": "query parameter uuid require"})
			return
		}
		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.JSON(http.StatusMethodNotAllowed, gin.H{"error_message": err.Error()})
			return
		}
		defer ws.Close()

		taskLog := TaskLog{
			DB:   db,
			UUID: execUUID,
		}
		for {
			status, log, err := taskLog.GetNewLog()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error_message": err.Error()})
				return
			}
			if err := ws.WriteMessage(websocket.TextMessage, []byte(log)); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error_message": err.Error()})
				return
			}
			if status == ProcessStatusFail || status == ProcessStatusSuccess {
				break
			}
			time.Sleep(2 * time.Second)
		}
	})

	if err := r.Run(appListen); err != nil {
		panic(err)
	}
	return
}
