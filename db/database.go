package db

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// Song представляет структуру песни.
type Song struct {
	ID          int
	GroupName   string
	SongName    string
	ReleaseDate string
	Text        string
	Link        string
}

// Repository представляет репозиторий для работы с базой данных.
type Repository struct {
	DB *sql.DB
}

// NewRepository создает новое подключение к базе данных.
func NewRepository(dsn string) (*Repository, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Printf("Error opening database: %v", err)
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err = db.Ping(); err != nil {
		log.Printf("Error connecting to database: %v", err)
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Successfully connected to the database")
	return &Repository{DB: db}, nil
}

// ValidateSong проверяет данные песни на корректность.
func (r *Repository) ValidateSong(song Song) error {
	// Валидация обязательных полей
	if song.GroupName == "" || song.SongName == "" {
		return fmt.Errorf("group_name and song_name are required")
	}

	// Валидация формата даты
	if _, err := time.Parse("2006-01-02", song.ReleaseDate); err != nil {
		return fmt.Errorf("invalid release_date format, expected YYYY-MM-DD")
	}

	// Валидация URL
	if _, err := url.ParseRequestURI(song.Link); err != nil {
		return fmt.Errorf("invalid URL format in link")
	}

	return nil
}

// GetSongs возвращает список песен по указанным фильтрам.
func (r *Repository) GetSongs(groupName, songName string, limit, offset int) ([]Song, error) {
	query := `
		SELECT id, group_name, song_name, release_date, text, link 
		FROM songs 
		WHERE group_name ILIKE $1 AND song_name ILIKE $2
		LIMIT $3 OFFSET $4
	`
	rows, err := r.DB.Query(query, "%"+groupName+"%", "%"+songName+"%", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query songs: %w", err)
	}
	defer rows.Close()

	var songs []Song
	for rows.Next() {
		var song Song
		if err := rows.Scan(&song.ID, &song.GroupName, &song.SongName, &song.ReleaseDate, &song.Text, &song.Link); err != nil {
			return nil, fmt.Errorf("failed to scan song: %w", err)
		}

		// Валидация данных после сканирования
		if song.GroupName == "" || song.SongName == "" {
			return nil, fmt.Errorf("invalid song data: missing group name or song name")
		}

		songs = append(songs, song)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return songs, nil
}

// GetSongText возвращает текст песни по ID.
func (r *Repository) GetSongText(songID int) (string, error) {
	if songID <= 0 {
		return "", fmt.Errorf("invalid song ID")
	}

	query := `SELECT text FROM songs WHERE id = $1`
	var text string
	err := r.DB.QueryRow(query, songID).Scan(&text)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("song with ID %d not found", songID)
		}
		return "", fmt.Errorf("failed to get song text: %w", err)
	}
	return text, nil
}

// DeleteSong удаляет песню по ID.
func (r *Repository) DeleteSong(songID int) error {
	if songID <= 0 {
		return fmt.Errorf("invalid song ID")
	}

	_, err := r.DB.Exec("DELETE FROM songs WHERE id = $1", songID)
	if err != nil {
		return fmt.Errorf("failed to delete song: %w", err)
	}
	return nil
}

// UpdateSong обновляет данные песни по ID.
func (r *Repository) UpdateSong(songID int, groupName, songName, releaseDate, text, link string) error {
	// Валидация входных данных
	if songID <= 0 {
		return fmt.Errorf("invalid song ID")
	}
	if groupName == "" || songName == "" {
		return fmt.Errorf("group_name and song_name are required")
	}
	if _, err := time.Parse("2006-01-02", releaseDate); err != nil {
		return fmt.Errorf("invalid release_date format, expected YYYY-MM-DD")
	}
	if _, err := url.ParseRequestURI(link); err != nil {
		return fmt.Errorf("invalid URL format in link")
	}

	query := `
		UPDATE songs 
		SET group_name = $1, song_name = $2, release_date = $3, text = $4, link = $5
		WHERE id = $6
	`
	_, err := r.DB.Exec(query, groupName, songName, releaseDate, text, link, songID)
	if err != nil {
		return fmt.Errorf("failed to update song: %w", err)
	}
	return nil
}

// CreateSong создает новую песню.
func (r *Repository) CreateSong(song Song) error {
	// Валидация данных
	if err := r.ValidateSong(song); err != nil {
		return err
	}

	query := `
		INSERT INTO songs (group_name, song_name, release_date, text, link) 
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.DB.Exec(query, song.GroupName, song.SongName, song.ReleaseDate, song.Text, song.Link)
	if err != nil {
		log.Printf("Error creating song: %v", err)
		return fmt.Errorf("failed to create song: %w", err)
	}
	return nil
}
