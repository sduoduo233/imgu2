package db

import "fmt"

type Storage struct {
	Id          int
	Name        string
	Type        string
	Config      string
	Enabled     bool
	AllowUpload bool
}

func StorageCreate(name string, storageType string, config string, enabled bool, allowUpload bool) (int, error) {
	r, err := DB.Exec("INSERT INTO storages(name, type, config, enabled, allow_upload) VALUES (?, ?, ?, ?, ?)", name, storageType, config, enabled, allowUpload)
	if err != nil {
		return 0, fmt.Errorf("db: %w", err)
	}
	n, err := r.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("db: %w", err)
	}
	return int(n), nil
}

func StorageFindAll() ([]Storage, error) {
	result := make([]Storage, 0)

	rows, err := DB.Query("SELECT id, name, type, config, enabled, allow_upload FROM storages ORDER BY id ASC")
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		s := Storage{}
		err = rows.Scan(&s.Id, &s.Name, &s.Type, &s.Config, &s.Enabled, &s.AllowUpload)
		if err != nil {
			return nil, fmt.Errorf("db: %w", err)
		}
		result = append(result, s)
	}

	return result, nil
}

func StorageFindById(id int) (*Storage, error) {
	var s Storage
	row := DB.QueryRow("SELECT id, name, type, config, enabled, allow_upload FROM storages WHERE id = ? LIMIT 1", id)
	err := row.Scan(&s.Id, &s.Name, &s.Type, &s.Config, &s.Enabled, &s.AllowUpload)
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}
	return &s, nil
}

func StorageSetEnabled(id int, enabled bool) error {
	_, err := DB.Exec("UPDATE storages SET enabled = ? WHERE id = ?", enabled, id)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}
	return nil
}

func StorageUpdate(id int, enabled bool, allowUpload bool, config string) error {
	_, err := DB.Exec("UPDATE storages SET enabled = ?, allow_upload = ?, config = ? WHERE id = ?", enabled, allowUpload, config, id)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}
	return nil
}

func StorageDelete(id int) error {
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	defer tx.Rollback()

	// check whether the storage driver is empty
	r := tx.QueryRow("SELECT COUNT(*) FROM images WHERE storage = ?", id)

	var cnt int
	err = r.Scan(&cnt)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	if cnt > 0 {
		return fmt.Errorf("db: storage delete: driver %d is not empty", id)
	}

	// delete the storage driver
	_, err = tx.Exec("DELETE FROM storages WHERE id = ?", id, false)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	// commit
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return nil
}
