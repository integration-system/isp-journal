package log

import (
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	Filename        string `schema:"Имя файла,путь до файла в который будут записываться логи"`
	MaxSizeMb       int    `schema:"Максимальный размер файла,ограничение по размеру файла после достижения которого логи будут записываться в новый файл"`
	RotateTimeoutMs int    `schema:"Время чередования файлов,ограничение по времени записи после достижения которого логи будут записываться в новый файл"`
	Compress        bool   `schema:"Не используется (сжатие логов,архивирует файлы в gzip)"`
	BufferSize      int    `schema:"Размер буфера,при указании разбивает данные и записывает их в файл по частям"`
}

func (c Config) GetFilename() string {
	name := ""
	if c.Filename == "" {
		name = os.Args[0] + ".log"
	} else {
		name = c.Filename
	}
	//if c.IsCompress() {
	//	name += ".gz"
	//}
	return name
}

func (c Config) GetMaxSizeInBytes() int64 {
	return int64(c.MaxSizeMb * 1024 * 1024)
}

func (c Config) GetRotateTimeout() time.Duration {
	return time.Duration(c.RotateTimeoutMs) * time.Millisecond
}

//func (c Config) IsCompress() bool {
//	return c.Compress
//}

func (c Config) IsBuffered() bool {
	return c.BufferSize > 0
}

func (c Config) GetBufferSize() int {
	return c.BufferSize
}

func (c Config) GetDirectory() string {
	return filepath.Dir(c.GetFilename())
}

func (c Config) GetFilePrefixAndExt() (string, string) {
	filename := filepath.Base(c.GetFilename())
	ext := filepath.Ext(filename)
	prefix := filename[:len(filename)-len(ext)] + "-"
	return prefix, ext
}
