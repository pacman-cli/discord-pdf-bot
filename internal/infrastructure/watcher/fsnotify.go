package watcher

import (
	"log/slog"
	"time"

	"github.com/fsnotify/fsnotify"
)

type FileWatcher struct {
	watcher  *fsnotify.Watcher
	debounce time.Duration
	onChange func()
}

func NewFileWatcher(folder string, onChange func()) (*FileWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	if err := w.Add(folder); err != nil {
		w.Close()
		return nil, err
	}

	fw := &FileWatcher{
		watcher:  w,
		debounce: 500 * time.Millisecond,
		onChange: onChange,
	}

	go fw.watch()

	slog.Info("Watching folder", "path", folder)
	return fw, nil
}

func (fw *FileWatcher) watch() {
	var timer *time.Timer

	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Create|fsnotify.Remove|fsnotify.Write) != 0 {
				slog.Debug("File change detected", "file", event.Name)
				if timer != nil {
					timer.Stop()
				}
				timer = time.AfterFunc(fw.debounce, fw.onChange)
			}
		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			slog.Error("Watcher error", "error", err)
		}
	}
}

func (fw *FileWatcher) Close() error {
	return fw.watcher.Close()
}
