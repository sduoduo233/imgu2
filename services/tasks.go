package services

import (
	"imgu2/db"
	"log/slog"
	"time"
)

func TaskStart() {
	// clean expired images
	taskRegister("clean images", time.Hour, func() error {
		images, err := db.ImageFindExpired()
		if err != nil {
			return err
		}

		for _, v := range images {
			// delete from storage
			err = Storage.DeleteFileFromDriver(v.StorageId, v.FileName)
			if err != nil {
				slog.Error("delete expired image", "storage", v.StorageId, "file name", v.FileName, "err", err)
				continue
			}

			// delete from database
			err = db.ImageDelete(v.Id)
			if err != nil {
				slog.Error("delete expired image", "storage", v.StorageId, "file name", v.FileName, "err", err)
			}
		}
		return nil
	})

	// clean expired sessions
	taskRegister("clean sessions", time.Hour, func() error {
		return db.SessionCleanExpired()
	})

}

func taskRegister(name string, d time.Duration, f func() error) {
	go func() {
		timer := time.NewTicker(d)
		for {
			slog.Info("executing task", "name", name)
			err := f()
			if err != nil {
				slog.Error("scheduled task", "name", name, "err", err)
			}

			<-timer.C
		}
	}()
}
