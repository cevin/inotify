package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

func main() {

	logger := log.New(os.Stdout, "", log.LstdFlags)

	// 解析命令行参数
	timeout := flag.Int("timeout", 0, "Timeout in seconds (exit if no events are detected within this time)")
	wait := flag.Int("wait", 0, "Wait in seconds (exit if no new events are detected after the first event)")
	events := flag.String("events", "create,write,remove,rename", "Comma-separated list of events to monitor (create, write, remove, rename)")
	exclude := flag.String("exclude", "", "Comma-separated list of file patterns to exclude")
	recursive := flag.Bool("r", false, "Recursively monitor all subdirectories")
	flag.Parse()

	// 获取目录路径
	if flag.NArg() == 0 {
		logger.Fatal("Directory path is required")
	}
	dir := flag.Arg(0)

	// 参数校验
	if *timeout < -1 || *wait < -1 {
		logger.Fatal("timeout and wait must be -1 (disabled) or non-negative")
	}

	// 解析事件参数
	eventList := strings.Split(*events, ",")
	eventMap := make(map[string]bool)
	for _, event := range eventList {
		eventMap[strings.TrimSpace(event)] = true
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		logger.Fatalf("Failed to get absolute path for directory %s: %v", dir, err)
	}

	excludeMaps := make([]*regexp.Regexp, 0)

	for _, path := range strings.Split(*exclude, ",") {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		path = filepath.Clean(path)
		regexPattern := "^" + regexp.QuoteMeta(filepath.Join(absDir, path))
		if strings.HasSuffix(path, string(filepath.Separator)) {
			regexPattern += ".*"
		}
		regex, err := regexp.Compile(regexPattern)
		if err != nil {
			logger.Fatalf("Failed to compile regex for exclude path %s: %v", path, err)
		}
		excludeMaps = append(excludeMaps, regex)
	}

	// 创建 Watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Fatalf("Failed to create watcher: %v", err)
	}
	defer func() {
		_ = watcher.Close()
	}()

	// 添加监听目录
	err = watcher.Add(dir)
	if err != nil {
		logger.Fatalf("Failed to add directory to watcher: %v", err)
	}
	if *recursive {
		err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			path, err = filepath.Abs(path)
			if err != nil {
				return err
			}
			if info.IsDir() {
				err = watcher.Add(path)
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			logger.Fatalln(err)
		}
	}

	doneChan := make(chan int)
	eventChan := make(chan fsnotify.Event, 10)
	timeoutTimer := time.NewTimer(0)
	waitTimer := time.NewTimer(0)

	// Immediately consume the timer. This line blocks until the timer fires,
	// and since the timer is set to 0, it will fire immediately, allowing the program
	// to continue execution after consuming the timer's event from its channel.
	<-timeoutTimer.C
	<-waitTimer.C

	if *timeout > 0 {
		timeoutTimer.Reset(time.Duration(*timeout) * time.Second)
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					logger.Println("Invalid watcher process")
					doneChan <- 255
					return
				}
				eventChan <- event
			case event := <-eventChan:
				detected := false

				absPath, err := filepath.Abs(event.Name)
				if err != nil {
					logger.Printf("Failed to get absolute path for event %s: %v\n", event.Name, err)
					doneChan <- 255
				}

				if (event.Op&fsnotify.Create == fsnotify.Create && eventMap["create"]) ||
					(event.Op&fsnotify.Write == fsnotify.Write && eventMap["write"]) ||
					(event.Op&fsnotify.Remove == fsnotify.Remove && eventMap["remove"]) ||
					(event.Op&fsnotify.Rename == fsnotify.Rename && eventMap["rename"]) {
					detected = true
				}

				if detected && *exclude != "" {
					for _, regex := range excludeMaps {
						if regex.MatchString(absPath) {
							detected = false
							break
						}
					}
				}

				if !detected {
					continue
				}

				logger.Printf("EVENT:%s %s\n", event.Op.String(), event.Name)

				if *timeout > 0 {
					timeoutTimer.Reset(time.Duration(*timeout) * time.Second)
				}
				if *wait > 0 {
					waitTimer.Reset(time.Duration(*wait) * time.Second)
				} else {
					waitTimer.Reset(0)
				}
			case <-timeoutTimer.C:
				doneChan <- 2
				return
			case <-waitTimer.C:
				doneChan <- 0
				return
			}
		}
	}()

	code := <-doneChan
	os.Exit(code)
}
