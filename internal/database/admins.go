package database // специализированные запросы или инфа по админам
// админы имеют доп.права в боте

import (
	"database/sql"
	"fmt"
	"workwavebot/internal/logger"
)

type Admins struct {
	ID   int64 // самим TG используется int64
	Name string
}

func (a *App) AddAdmin(id int64, name string) error {
	query := `INSERT INTO admins (id, name) VALUES ($1, $2)`
	_, err := a.db.Exec(query, id, name)
	if err != nil {
		logger.ErrLog.Printf("ошибка AddAdmin: %v", err)
		return fmt.Errorf("add admin: %w", err)
	}
	return nil
}

func (a *App) DeleteAdmin(id int64) error {
	query := `DELETE FROM admins WHERE id = $1`
	result, err := a.db.Exec(query, id)
	if err != nil {
		logger.ErrLog.Printf("ошибка DeleteAdmin: %v", err)
		return fmt.Errorf("delete admin: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("админ с ID %d не найден", id)
	}
	return nil
}

func (a *App) GetAdmins() ([]Admins, error) {
	query := `SELECT id, name FROM admins ORDER BY id ASC`
	rows, err := a.db.Query(query)
	if err != nil {
		logger.ErrLog.Printf("ошибка GetAdmins: %v", err)
		return nil, fmt.Errorf("get admins: %w", err)
	}
	defer rows.Close()

	var admins []Admins
	for rows.Next() {
		var ad Admins
		if err := rows.Scan(&ad.ID, &ad.Name); err != nil {
			logger.ErrLog.Printf("ошибка scan admin: %v", err)
			continue
		}
		admins = append(admins, ad)
	}
	return admins, nil
}

func (a *App) CheckAdmin(userID int64) bool {
	query := `SELECT EXISTS (SELECT 1 FROM admins WHERE id = $1)`
	var exists bool
	err := a.db.QueryRow(query, userID).Scan(&exists)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.ErrLog.Printf("ошибка CheckAdmin %d: %v", userID, err)
		}
		return false
	}
	return exists
}
