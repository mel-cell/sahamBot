package db

import (
	"database/sql"
	"fmt"
	"log"
	
	_ "github.com/mattn/go-sqlite3" // Import driver sqlite3
)

type Repository struct {
	DB *sql.DB
}

// Watchlist Struct
type Watchlist struct {
	ID        int
	UserID    int64
	Symbol    string
	CreatedAt string
}

// Init Database (Create Table if not exists)
func NewRepository(dbPath string) (*Repository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	
	// Create Table
	query := `
	CREATE TABLE IF NOT EXISTS watchlists (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		symbol TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(user_id, symbol)
	);`
	
	if _, err := db.Exec(query); err != nil {
		return nil, fmt.Errorf("failed to create table: %v", err)
	}
	
	return &Repository{DB: db}, nil
}

// Add Watchlist Item
func (r *Repository) AddWatch(userID int64, symbol string) error {
	query := "INSERT INTO watchlists (user_id, symbol) VALUES (?, ?)"
	_, err := r.DB.Exec(query, userID, symbol)
	return err
}

// Remove Watchlist Item
func (r *Repository) RemoveWatch(userID int64, symbol string) error {
	query := "DELETE FROM watchlists WHERE user_id = ? AND symbol = ?"
	_, err := r.DB.Exec(query, userID, symbol)
	return err
}

// Get User Watchlist
func (r *Repository) GetUserWatchlist(userID int64) ([]string, error) {
	query := "SELECT symbol FROM watchlists WHERE user_id = ?"
	rows, err := r.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var symbols []string
	for rows.Next() {
		var sym string
		if err := rows.Scan(&sym); err != nil {
			return nil, err
		}
		symbols = append(symbols, sym)
	}
	return symbols, nil
}

// Get All Watchlist (For Scheduler Check)
// Returns Map[UserID] -> []Symbol
func (r *Repository) GetAllUniqueUserSymbols() (map[int64][]string, error) {
	query := "SELECT user_id, symbol FROM watchlists"
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64][]string)
	for rows.Next() {
		var uid int64
		var sym string
		if err := rows.Scan(&uid, &sym); err != nil {
			continue
		}
		result[uid] = append(result[uid], sym)
	}
	return result, nil
}
