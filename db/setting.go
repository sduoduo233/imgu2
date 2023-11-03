package db

import "fmt"

func SettingFind(key string) (string, error) {
	var v string
	row := DB.QueryRow("SELECT value FROM settings WHERE key = ? LIMIT 1", key)
	err := row.Scan(&v)
	if err != nil {
		return "", fmt.Errorf("db: %w", err)
	}
	return v, nil
}

func SettingFindAll() (map[string]string, error) {
	m := make(map[string]string)
	rows, err := DB.Query("SELECT key, value FROM settings")
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var key string
		var value string
		err := rows.Scan(&key, &value)
		if err != nil {
			return nil, fmt.Errorf("db: %w", err)
		}
		m[key] = value
	}

	return m, nil
}

func SetttingUpdate(key, value string) error {
	_, err := DB.Exec("UPDATE settings SET value = ? WHERE key = ?", value, key)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}
	return nil
}
