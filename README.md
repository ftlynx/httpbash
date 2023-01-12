# httpbash
用于包装shell脚本，让其它机器可以通过http执行本机的shell脚本

## require
需要自己安装 mysql

## run

```
go build .
cp config.default config # 修改 datasource 配置
./httpbash 
```

## 执行命令
### 参数说明
- cmd 执行命令 必须
- timeout_minute 超时时间，默认60，单位是分钟。非必须
- created_user 创建人，非必须
- uuid 非必须，用户可以自己传入uuid，当需要传递配置文件是必须指定，需要用这个id生成文件名
- config_file.base64_content": "" 非必须, 用于一些命令需要传输配置文件。文件存放路径run/config/{uuid}
```
# 普通命令
curl -XPOST -H "x-api-auth:changeme" http://127.0.0.1:8080/command -d '
{
    "cmd": "ping www.baidu.com",
    "timeout_minute": 1,
    "created_user": "xx@example.com"
}'
{"exec_uuid":"41f435d7-d4a8-4137-af8c-4e228fff26bb"}

#需要传递配置文件
echo -n "abc" | base64  # 将文件base64后放到 config_file.base64_content 字段上面
curl -XPOST -H "x-api-auth:changeme" http://127.0.0.1:8080/command -d '
{
    "uuid":"41f435d7-d4a8-4137-af8c-4e228fff26bc",
    "cmd": "cat run/config/41f435d7-d4a8-4137-af8c-4e228fff26bc",
    "timeout_minute": 1,
    "created_user": "xx@example.com",
    "config_file":{
        "base64_content": "YWJj"
    }
}'
{"exec_uuid":"41f435d7-d4a8-4137-af8c-4e228fff26bc"}
```

```
curl -s -H "x-api-auth:changeme" http://127.0.0.1:8080/command?uuid=41f435d7-d4a8-4137-af8c-4e228fff26bb | python3 -m json.tool
{
    "uuid": "41f435d7-d4a8-4137-af8c-4e228fff26bb",
    "cmd_string": "/sbin/ping www.baidu.com",
    "stdout": "PING www.a.shifen.com (111.206.208.134): 56 data bytes\n64 bytes from 111.206.208.134: icmp_seq=0 ttl=48 time=44.749 ms\n64 bytes from 111.206.208.134: icmp_seq=1 ttl=48 time=54.050 ms\n64 bytes from 111.206.208.134: icmp_seq=2 ttl=48 time=55.768 ms\n64 bytes from 111.206.208.134: icmp_seq=3 ttl=48 time=47.418 ms\n64 bytes from 111.206.208.134: icmp_seq=4 ttl=48 time=42.322 ms\n64 bytes from 111.206.208.134: icmp_seq=5 ttl=48 time=45.283 ms\n64 bytes from 111.206.208.134: icmp_seq=6 ttl=48 time=48.437 ms\n64 bytes from 111.206.208.134: icmp_seq=7 ttl=48 time=41.499 ms\n64 bytes from 111.206.208.134: icmp_seq=8 ttl=48 time=42.855 ms\n64 bytes from 111.206.208.134: icmp_seq=9 ttl=48 time=42.221 ms\n64 bytes from 111.206.208.134: icmp_seq=10 ttl=48 time=43.430 ms\n64 bytes from 111.206.208.134: icmp_seq=11 ttl=48 time=43.125 ms\n64 bytes from 111.206.208.134: icmp_seq=12 ttl=48 time=60.551 ms\n64 bytes from 111.206.208.134: icmp_seq=13 ttl=48 time=46.066 ms\n64 bytes from 111.206.208.134: icmp_seq=14 ttl=48 time=47.909 ms\n64 bytes from 111.206.208.134: icmp_seq=15 ttl=48 time=43.357 ms\n64 bytes from 111.206.208.134: icmp_seq=16 ttl=48 time=45.696 ms\n64 bytes from 111.206.208.134: icmp_seq=17 ttl=48 time=42.205 ms\n64 bytes from 111.206.208.134: icmp_seq=18 ttl=48 time=43.904 ms\n64 bytes from 111.206.208.134: icmp_seq=19 ttl=48 time=44.077 ms\n64 bytes from 111.206.208.134: icmp_seq=20 ttl=48 time=42.942 ms\n64 bytes from 111.206.208.134: icmp_seq=21 ttl=48 time=44.298 ms\n64 bytes from 111.206.208.134: icmp_seq=22 ttl=48 time=42.034 ms\n64 bytes from 111.206.208.134: icmp_seq=23 ttl=48 time=43.649 ms\n64 bytes from 111.206.208.134: icmp_seq=24 ttl=48 time=43.890 ms\n64 bytes from 111.206.208.134: icmp_seq=25 ttl=48 time=42.970 ms\n64 bytes from 111.206.208.134: icmp_seq=26 ttl=48 time=50.603 ms\n64 bytes from 111.206.208.134: icmp_seq=27 ttl=48 time=57.424 ms\n64 bytes from 111.206.208.134: icmp_seq=28 ttl=48 time=45.747 ms\n64 bytes from 111.206.208.134: icmp_seq=29 ttl=48 time=44.980 ms\n64 bytes from 111.206.208.134: icmp_seq=30 ttl=48 time=43.423 ms\n64 bytes from 111.206.208.134: icmp_seq=31 ttl=48 time=52.314 ms\n64 bytes from 111.206.208.134: icmp_seq=32 ttl=48 time=46.739 ms\n64 bytes from 111.206.208.134: icmp_seq=33 ttl=48 time=48.679 ms\n64 bytes from 111.206.208.134: icmp_seq=34 ttl=48 time=43.210 ms\n64 bytes from 111.206.208.134: icmp_seq=35 ttl=48 time=41.960 ms\n64 bytes from 111.206.208.134: icmp_seq=36 ttl=48 time=42.072 ms\n64 bytes from 111.206.208.134: icmp_seq=37 ttl=48 time=43.046 ms\n64 bytes from 111.206.208.134: icmp_seq=38 ttl=48 time=41.163 ms\n64 bytes from 111.206.208.134: icmp_seq=39 ttl=48 time=41.427 ms\n64 bytes from 111.206.208.134: icmp_seq=40 ttl=48 time=43.156 ms\n64 bytes from 111.206.208.134: icmp_seq=41 ttl=48 time=51.481 ms\n64 bytes from 111.206.208.134: icmp_seq=42 ttl=48 time=43.678 ms\n64 bytes from 111.206.208.134: icmp_seq=43 ttl=48 time=42.436 ms\n64 bytes from 111.206.208.134: icmp_seq=44 ttl=48 time=43.306 ms\n64 bytes from 111.206.208.134: icmp_seq=45 ttl=48 time=42.223 ms\n64 bytes from 111.206.208.134: icmp_seq=46 ttl=48 time=41.148 ms\n64 bytes from 111.206.208.134: icmp_seq=47 ttl=48 time=41.829 ms\n64 bytes from 111.206.208.134: icmp_seq=48 ttl=48 time=42.527 ms\n64 bytes from 111.206.208.134: icmp_seq=49 ttl=48 time=41.818 ms\n64 bytes from 111.206.208.134: icmp_seq=50 ttl=48 time=43.246 ms\n64 bytes from 111.206.208.134: icmp_seq=51 ttl=48 time=44.290 ms\n64 bytes from 111.206.208.134: icmp_seq=52 ttl=48 time=43.395 ms\n64 bytes from 111.206.208.134: icmp_seq=53 ttl=48 time=41.397 ms\n64 bytes from 111.206.208.134: icmp_seq=54 ttl=48 time=45.780 ms\n64 bytes from 111.206.208.134: icmp_seq=55 ttl=48 time=42.038 ms\n64 bytes from 111.206.208.134: icmp_seq=56 ttl=48 time=51.113 ms\n64 bytes from 111.206.208.134: icmp_seq=57 ttl=48 time=41.923 ms\n64 bytes from 111.206.208.134: icmp_seq=58 ttl=48 time=42.518 ms\n64 bytes from 111.206.208.134: icmp_seq=59 ttl=48 time=40.948 ms\n",
    "process_status": 1,
    "process_fail_log": "signal: killed",
    "created_user": "xx@example.com",
    "created_at": 1673323527536,
    "updated_at": 1673323587536
}
```

## 状态(process_status)
- 1 执行失败
- 2 执行中
- 3 执行成功
    