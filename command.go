package main

import (
	"bufio"
	"context"
	"fmt"
	"gorm.io/gorm"
	"log"
	"os"
	"os/exec"
	"time"
)

type Command struct {
	Name        string
	Args        []string
	Timout      time.Duration
	UUID        string
	CreatedUser string
	DB          *gorm.DB
}

func (c *Command) Exec() {
	log.SetOutput(os.Stdout)
	log.SetPrefix(fmt.Sprintf("%s\t", c.UUID))

	ctx, cancel := context.WithTimeout(context.Background(), c.Timout)
	defer cancel()

	cmd := exec.CommandContext(ctx, c.Name, c.Args...)
	log.Printf("[%s] exec begin\n", cmd.String())
	if err := c.DB.Create(&Task{
		UUID:          c.UUID,
		CmdString:     cmd.String(),
		CreatedUser:   c.CreatedUser,
		ProcessStatus: ProcessStatusRunning,
	}).Error; err != nil {
		log.Printf("DB.Create err: %s", err)
		return
	}

	stdout, err := cmd.StdoutPipe() // 获取标准输出
	if err != nil {
		log.Printf("cmd.StdoutPipe err: %s\n", err.Error())
		return
	}
	cmd.Stderr = cmd.Stdout // 错误输出写到标准输出

	if err := cmd.Start(); err != nil {
		log.Printf("cmd.Start err: %s\n", err.Error())
		if err := c.DB.Model(&Task{}).Where("uuid=?", c.UUID).Updates(Task{
			ProcessStatus: ProcessStatusFail,
			ProcessLog:    err.Error(),
		}).Error; err != nil {
			log.Printf("DB.Update(ProcessStatusFail) by cmd.Start err: %s\n", err.Error())
			return
		}
		return
	}

	result := ""
	br := bufio.NewReader(stdout)
	for {
		// todo 日志输出过快会导致数据库 update 太频繁
		b, err := br.ReadBytes('\n')
		if err != nil && len(b) == 0 {
			break
		}
		result = result + string(b)
		if err := c.DB.Model(&Task{}).Where("uuid=?", c.UUID).Updates(Task{
			ProcessLog: result,
		}).Error; err != nil {
			log.Printf("DB.Update(stdout) err: %s\n", err.Error())
			return
		}
	}

	// wait 中 context 处理了超时，不需要在处理了
	if err := cmd.Wait(); err != nil {
		log.Printf("cmd.Wait err: %s\n", err.Error())
		if err := c.DB.Model(&Task{}).Where("uuid=?", c.UUID).Updates(Task{
			ProcessStatus: ProcessStatusFail,
			ProcessLog:    result + err.Error(),
		}).Error; err != nil {
			log.Printf("DB.Update(ProcessStatusFail) by cmd.Wait err: %s\n", err.Error())
			return
		}
		return
	}

	if err := c.DB.Model(&Task{}).Where("uuid=?", c.UUID).Updates(Task{
		ProcessStatus: ProcessStatusSuccess,
	}).Error; err != nil {
		log.Printf("DB.Update(ProcessStatusSuccess) err: %s\n", err.Error())
		return
	}

	log.Printf("[%s] exec success\n", cmd.String())
	return
}
