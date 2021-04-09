package slog

import (
	"fmt"
	"os"
	"sync"
	"time"
)

//日志等级
type LOGLevel int

const (
	DEBUG LOGLevel = iota
	WARNING
	ERROR
	INFO
)

var (
	vLevelName = [...]string{"DEBUG", "WARNING", "ERROR", "INFO"}
)

type sLOGWriteItem struct {
	level   LOGLevel
	wfile   bool
	wtime   time.Time
	wname   string
	content string
}

type sLOGWriterService struct {
	running bool                //服务状态
	dir     string              //日志目录
	level   LOGLevel            //日志写入等级
	items   chan *sLOGWriteItem //日志记录列表
	status  chan bool
	files   map[string]*os.File //文件列表
	buffer  []byte
	wg      sync.WaitGroup
}

var sService = &sLOGWriterService{
	running: false,
	dir:     "./dir",//options.LOG_DIR,
	level:   DEBUG,
	items:   make(chan *sLOGWriteItem, 1024),//options.LOG_ALLOC_ITEM),
	files:   make(map[string]*os.File),
	status:  make(chan bool),
	buffer:  make([]byte, 2048),//options.LOG_BUFFER_SIZE),
}

func fmtWriteBuffer(buf *[]byte, level LOGLevel, writetime time.Time, content string) {
	vars := time.Date(writetime.Year(), writetime.Month(), writetime.Day(), writetime.Hour(),
		writetime.Minute(), writetime.Second(), 0, time.Local)
	varstime := vars.Format("2006-01-02 15:04:05")

	*buf = append(*buf, varstime...)
	*buf = append(*buf, ' ')

	*buf = append(*buf, '[')
	*buf = append(*buf, vLevelName[int(level)]...)
	*buf = append(*buf, ']', ' ')

	*buf = append(*buf, content...)
	*buf = append(*buf, '\n')
}

func run() {
	defer sService.wg.Done()
	sService.wg.Add(1)
	sService.running = true
	for {
		select {
		case data, ok := <-sService.items:
			if data == nil || ok == false {
				return
			}
			sService.buffer = sService.buffer[:0]
			fmtWriteBuffer(&sService.buffer, data.level, data.wtime, data.content)
			if data.wfile {
				sService.writeFile(data.wname)
			}
			os.Stdout.Write(sService.buffer)
		case status, ok := <-sService.status:
			if status == false && ok {
				return
			}
		}
	}
}

func (service *sLOGWriterService) writeFile(writename string) {
	defer func() {
		if e := recover(); e != nil {
		}
	}()
	fWriter, ok := service.files[writename]
	if !ok {
		var err error
		sAbsFileName := service.dir + "/" + writename + ".log"
		fWriter, err = os.OpenFile(sAbsFileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
		if err == nil {
			service.files[writename] = fWriter
		}
	}
	fWriter.Write(service.buffer)
}

func (service *sLOGWriterService) write(level LOGLevel, writename string, writefile bool, format string, args ...interface{}) {
	if service.level > level { // || !service.service
		return
	}
	if len(args) > 0 {
		format = fmt.Sprintf(format, args...)
	}
	data := &sLOGWriteItem{
		level:   level,
		wfile:   writefile,
		wname:   writename,
		wtime:   time.Now(),
		content: format,
	}
	service.items <- data
}

//接口：写日志
func WriteLog(lev LOGLevel, desc string, format string, args ...interface{}) {
	//wfile := (lev >= INFO)
	wfile := false
	sService.write(lev, desc, wfile, format, args...)
}

func SetOption(dir string, writelv LOGLevel) {
	sService.dir = dir
	sService.level = writelv
}

func Run() {
	if sService.running == false {
		go run()
	}
}

func Close() {
	defer func() {
		for _, handler := range sService.files {
			handler.Sync()
			handler.Close()
		}
	}()
	sService.running = false
	sService.status <- false
	sService.wg.Wait()
}
