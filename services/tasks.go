package services

import (
	"imgu2/db"
	"log/slog"
	"time"
)

var taskStopChan = make(chan int)

func TaskStop() {
	close(taskStopChan)
}

func TaskStart() {
	// clean expired images
	taskRegister("clean images", time.Hour, func() error {
		images, err := db.ImageFindExpired()
		if err != nil {
			return err
		}

		for _, v := range images {
			// delete from storage
			err = Storage.DeleteFileFromDriver(v.StorageId, v.InternalName)
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

	// set expired user group to default
	taskRegister("reset user group", time.Hour*6, func() error {
		id, err := Setting.DefaultGroupRegistered()
		if err != nil {
			return err
		}

		slog.Info("reset user group", "id", id)

		err = db.UserResetExpiredGroup(id)
		if err != nil {
			return err
		}

		return nil
	})

}

func taskRegister(name string, d time.Duration, f func() error) {
	go func() {
		timer := time.NewTicker(d)

	Exit:
		for {
			select {
			case <-timer.C:
				slog.Info("executing task", "name", name)
				err := f()
				if err != nil {
					slog.Error("scheduled task", "name", name, "err", err)
				}
			case <-taskStopChan:
				slog.Debug("task stop", "name", name)
				timer.Stop()
				break Exit
			}

		}

	}()
}
