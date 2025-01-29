package connection

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/noctusha/music/models"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

// Repository provides methods to interact with the database.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new Repository with a database connection.
func NewRepository() (*Repository, error) {
	connStr := os.Getenv("POSTGRES_CONN")
	if connStr == "" {
		return nil, fmt.Errorf("POSTGRES_CONN environment variable not set")
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening connection to postgres: %v", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error pinging postgres: %v", err)
	}

	return &Repository{db: db}, nil
}

// Close closes the database connection.
func (r *Repository) Close() {
	r.db.Close()
}

// SongList retrieves a list of songs from the database with optional filters and pagination.
func (r *Repository) SongList(group, name, releaseDate, text, link string, limit, offset int) ([]models.Song, error) {
	var songs []models.Song
	var (
		rows         *sql.Rows
		err          error
		params       []interface{}
		whereClauses []string
	)

	if limit == 0 {
		limit = 25
	}

	query := `
SELECT
	songs.id,
	songs.name,
	songs.group_id
FROM
    songs
JOIN
	song_details
ON
	songs.id = song_details.song_id`

	if group != "" {
		whereClauses = append(whereClauses, "songs.group_id = (SELECT id FROM groups WHERE name ILIKE $"+fmt.Sprint(len(params)+1)+")")
		params = append(params, "%"+group+"%")
	}

	if name != "" {
		whereClauses = append(whereClauses, "songs.name ILIKE $"+fmt.Sprint(len(params)+1))
		params = append(params, "%"+name+"%")
	}

	if releaseDate != "" {
		whereClauses = append(whereClauses, "song_details.release_date = $"+fmt.Sprint(len(params)+1))
		params = append(params, releaseDate)
	}

	if text != "" {
		whereClauses = append(whereClauses, "song_details.text ILIKE $"+fmt.Sprint(len(params)+1))
		params = append(params, "%"+text+"%")
	}

	if link != "" {
		whereClauses = append(whereClauses, "song_details.link = $"+fmt.Sprint(len(params)+1))
		params = append(params, link)
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	query += `
ORDER BY
	songs.name
LIMIT
	$` + fmt.Sprint(len(params)+1) + `
OFFSET
	$` + fmt.Sprint(len(params)+2)

	params = append(params, limit, offset)

	rows, err = r.db.Query(query, params...)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		song := models.Song{}
		err = rows.Scan(&song.ID, &song.Name, &song.GroupID)
		if err != nil {
			return nil, fmt.Errorf("error scanning song: %v", err)
		}
		songs = append(songs, song)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("rows iteration error: %v", err)
	}

	if len(songs) == 0 {
		return songs, nil
	}

	return songs, nil
}

// TextListByID retrieves the text of a song by its ID.
func (r *Repository) TextListByID(id string) (string, bool, error) {
	var text string

	err := r.db.QueryRow("SELECT text FROM song_details WHERE song_id = $1", id).Scan(&text)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("error scanning song: %v", err)
	}
	return text, true, nil
}

// SongDelete deletes a song from the database by its ID.
func (r *Repository) SongDelete(songID string) error {
	_, err := r.db.Exec("DELETE FROM songs WHERE id = $1", songID)
	if err != nil {
		return fmt.Errorf("error deleting song: %v", err)
	}
	return nil
}

// GetGroupID retrieves the ID of a group by its name.
func (r *Repository) GetGroupID(group string) (int, error) {
	var id int

	err := r.db.QueryRow("SELECT id FROM groups WHERE name = $1", group).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("error scanning group: %v", err)
	}
	return id, nil
}

// GetSongByID retrieves a song by its ID.
func (r *Repository) GetSongByID(songID string) (*models.Song, error) {
	var song models.Song

	err := r.db.QueryRow("SELECT id, name, group_id FROM songs WHERE id = $1", songID).Scan(&song.ID, &song.Name, &song.GroupID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error scanning song: %v", err)
	}

	return &song, nil
}

// NewGroup creates a new group in the database.
func (r *Repository) NewGroup(name string) (int, error) {
	var id int
	err := r.db.QueryRow("INSERT INTO groups(name) VALUES ($1) RETURNING id", name).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("error inserting group: %v", err)
	}

	return id, nil
}

// GetSongDetailsByID retrieves song details by song ID.
func (r *Repository) GetSongDetailsByID(songID string) (*models.SongDetails, error) {
	var songDetails models.SongDetails

	err := r.db.QueryRow("SELECT id, song_id, release_date, text, link FROM song_details WHERE song_id = $1", songID).Scan(&songDetails.ID, &songDetails.SongID, &songDetails.ReleaseDate, &songDetails.Text, &songDetails.Link)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error scanning song: %v", err)
	}

	return &songDetails, nil
}

// UpdateSong updates a song and its details in the database.
func (r *Repository) UpdateSong(song *models.Song, songDetails *models.SongDetails) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	_, err = tx.Exec(`UPDATE songs SET name = $1, group_id = $2 WHERE id = $3`, song.Name, song.GroupID, song.ID)
	if err != nil {
		return fmt.Errorf("error updating song: %v", err)
	}

	_, err = tx.Exec(`UPDATE song_details SET release_date = $1, text = $2, link = $3 WHERE song_id = $4`, songDetails.ReleaseDate, songDetails.Text, songDetails.Link, songDetails.SongID)
	if err != nil {
		return fmt.Errorf("error updating song_details: %v", err)
	}

	return nil
}

// CreateSongWithDetails creates a new song and its details in the database.
func (r *Repository) CreateSongWithDetails(song models.Song, details models.SongDetails) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	var songID int
	err = tx.QueryRow(`INSERT INTO songs (name, group_id) VALUES ($1, $2) RETURNING id`, song.Name, song.GroupID).Scan(&songID)
	if err != nil {
		return fmt.Errorf("failed to insert song: %v", err)
	}

	_, err = tx.Exec(`INSERT INTO song_details (song_id, release_date, text, link) VALUES ($1, $2, $3, $4)`,
		songID, details.ReleaseDate, details.Text, details.Link)
	if err != nil {
		return fmt.Errorf("failed to insert song details: %v", err)
	}

	return nil
}
