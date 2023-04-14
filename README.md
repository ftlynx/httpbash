# httpbash
用于包装shell脚本，让其它机器可以通过http执行本机的shell脚本，可以通过websocket查看日志


## run

```
bash build.sh
cp config.yaml.default config.yaml 
./httpbash run
```

## 执行命令
### 参数说明
- cmd 执行命令 必须，如何需要传递配置文件，需要使用{{.config}} 占用位置
- timeout_minute 超时时间，默认60，单位是分钟。非必须
- created_user 创建人，非必须
- task_id 必须，任务ID
- config_file.base64_content": "" 非必须, 用于一些命令需要传输配置文件
```
# 普通命令
curl -XPOST -H "x-api-auth:changeme" http://127.0.0.1:8080/v1/command -d '
{
    "task_id":"aaaa-bbbb-cccc-dddd-eeef",
    "cmd": "ping www.baidu.com",
    "timeout_minute": 1,
    "created_user": "xx@example.com"
}'
{
    "success": true,
    "data": {
        "http_api_endpoint": "http://127.0.0.1:9191/v0/command/log?task_id=aaaa-bbbb-cccc-dddd-eeef&job_id=1681382803555",
        "ws_api_endpoint": "http://127.0.0.1:9191/v0/command/log/ws?task_id=aaaa-bbbb-cccc-dddd-eeef&job_id=1681382803555",
        "ws_html_endpoint": "http://127.0.0.1:9191/v0/console/?task_id=aaaa-bbbb-cccc-dddd-eeef&job_id=1681382803555"
    }
}

#需要传递配置文件
#配置文件使用模板变量 {{.config}} 占位置
echo -n "abc" | base64  # 将文件base64后放到 config_file.base64_content 字段上面
curl -XPOST -H "x-api-auth:changeme" http://127.0.0.1:8080/v1/command -d '
{
    "task_id":"aaaa-bbbb-cccc-dddd-eeef",
    "cmd": "cat {{.config}}",
    "timeout_minute": 1,
    "created_user": "xx@example.com",
    "config_file":{
        "base64_content": "YWJj"
    }
}'
{
    "success": true,
    "data": {
        "http_api_endpoint": "http://127.0.0.1:9191/v0/command/log?task_id=aaaa-bbbb-cccc-dddd-eeef&job_id=1681382803555",
        "ws_api_endpoint": "http://127.0.0.1:9191/v0/command/log/ws?task_id=aaaa-bbbb-cccc-dddd-eeef&job_id=1681382803555",
        "ws_html_endpoint": "http://127.0.0.1:9191/v0/console/?task_id=aaaa-bbbb-cccc-dddd-eeef&job_id=1681382803555"
    }
}
```

### 返回结果说明
```
"data": {
        "http_api_endpoint": "http://127.0.0.1:9191/v0/command/log?task_id=aaaa-bbbb-cccc-dddd-eeef&job_id=1681382803555",
        "ws_api_endpoint": "http://127.0.0.1:9191/v0/command/log/ws?task_id=aaaa-bbbb-cccc-dddd-eeef&job_id=1681382803555",
        "ws_html_endpoint": "http://127.0.0.1:9191/v0/console/?task_id=aaaa-bbbb-cccc-dddd-eeef&job_id=1681382803555"
}
```
- http_api_endpoint 表示通过http api获取命令运行日志。
- ws_api_endpoint 表示通过ws api获取命令运行日志。
- ws_html_endpoint 一个简单的html页面，自带websocket。


    