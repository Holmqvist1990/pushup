package build

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// FIXME(paulsmith): package global db conn
var DB *sql.DB

// FIXME(paulsmith): relying on init() is not great from an app lifecycle POV
func init() {
	dbPath := "./mypushupapp.db"
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(fmt.Errorf("opening SQLite db %s: %w", dbPath, err))
	}
	if err := initDb(db); err != nil {
		panic(fmt.Errorf("initializing SQLite db: %w", err))
	}
	DB = db
}

var createTable = `
CREATE TABLE IF NOT EXISTS [albums] (
	[id] INTEGER PRIMARY KEY,
	[artist] TEXT,
	[title] TEXT,
	[released] INTEGER,
	[length] INTEGER
);
`

func initDb(db *sql.DB) error {
	if _, err := db.Exec(createTable); err != nil {
		return fmt.Errorf("creating albums table: %w", err)
	}
	return nil
}

type album struct {
	id       int
	artist   string
	title    string
	released int
	length   int
}

var insertAlbumRow = `
INSERT INTO [albums] ([artist], [title], [released], [length])
VALUES (?, ?, ?, ?)
RETURNING [id]
`

func addAlbum(db *sql.DB, a *album) error {
	args := []any{
		a.artist,
		a.title,
		a.released,
		a.length,
	}
	if err := db.QueryRow(insertAlbumRow, args...).Scan(&a.id); err != nil {
		return fmt.Errorf("inserting album: %w", err)
	}
	return nil
}

var selectAlbumById = `
SELECT [artist], [title], [released], [length]
FROM [albums]
WHERE [id] = ?
`

func getAlbumById(db *sql.DB, id int) (*album, error) {
	a := album{id: id}
	dest := []any{
		&a.artist,
		&a.title,
		&a.released,
		&a.length,
	}
	if err := db.QueryRow(selectAlbumById, id).Scan(dest...); err != nil {
		return nil, fmt.Errorf("getting album by ID: %w", err)
	}
	return &a, nil
}

var selectAlbums = `
SELECT [id], [artist], [title], [released], [length]
FROM [albums]
ORDER BY id
`

func getAlbums(db *sql.DB, limit, offset int) ([]*album, error) {
	query := selectAlbums
	var args []any
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}
	if offset > 0 {
		query += " OFFSET ?"
		args = append(args, offset)
	}
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("getting albums: %v", err)
	}
	defer rows.Close()
	var albums []*album
	for rows.Next() {
		var a album
		dest := []any{
			&a.id,
			&a.artist,
			&a.title,
			&a.released,
			&a.length,
		}
		if err := rows.Scan(dest...); err != nil {
			return nil, fmt.Errorf("getting albums, scanning row: %w", err)
		}
		albums = append(albums, &a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("getting albums, scan: %w", err)
	}
	return albums, nil
}

var updateAlbum = `
UPDATE [albums]
SET
	[artist] = ?,
	[title] = ?,
	[released] = ?,
	[length] = ?
WHERE [id] = ?
`

func editAlbum(db *sql.DB, id int, a *album) error {
	args := []any{
		&a.artist,
		&a.title,
		&a.released,
		&a.length,
		id,
	}
	if _, err := db.Exec(updateAlbum, args...); err != nil {
		return fmt.Errorf("updating album: %v", err)
	}
	return nil
}

var deleteAlbum_ = `DELETE FROM [albums] WHERE [id] = ?`

func deleteAlbum(db *sql.DB, id int) error {
	if _, err := db.Exec(deleteAlbum_, id); err != nil {
		return fmt.Errorf("deleting album: %v", err)
	}
	return nil
}
