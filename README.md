# inoitfy

`inoitfy` 是一个用 Go 语言编写的文件监控工具，和 `inotify-tools` 类似，它能够在指定时间范围内对一批文件事件进行监听。此工具借助
`fsnotify` 库实现文件系统事件的监听，可根据用户设定的参数对文件的创建、写入、删除和重命名等操作进行灵活监控。

## 安装

在安装该工具前，请确保你已经安装了 Go 环境。使用以下命令来安装 `inoitfy`：

```bash

go get github.com/cevin/inoitfy
```

## 使用方法

基本命令格式

```text
inoitfy [参数] <目录路径>
```

### 参数说明

| 参数         | 说明                                                      | 默认值                        |          
|:-----------|:--------------------------------------------------------|:---------------------------|
| `-timeout` | 超时时间（秒）。若在此时间内未检测到任何事件，程序将退出。                           | 0                          |
| `-wait`    | 等待时间（秒）。在首次检测到事件后，若在此时间内没有新事件发生，程序将退出。                  | 0                          |
| `-events`  | 以逗号分隔的事件列表，指定要监控的事件类型，支持 create、write、remove 和 rename。	 | create,write,remove,rename |
| `-exclude` | 以逗号分隔的文件模式列表，用于排除不需要监控的文件，支持使用通配符。	                     | 空                          |

## Example

### 监控指定目录下的所有默认事件，设置超时时间为 60 秒

```bash
inoitfy -timeout 60 /path/to/directory
```

### 仅监控文件创建和写入事件，排除所有 .log 文件

```bash
inoitfy -events create,write -exclude *.log /path/to/directory
```

### 首次检测到事件后，等待 30 秒，若没有新事件则退出

```bash
inoitfy -wait 30 /path/to/directory
```

# LICENSE

MIT License