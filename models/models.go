package models

// Group represents a musical group or artist.
type Group struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Song represents a song.
type Song struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	GroupID int    `json:"group_id"`
}

// SongDetails contains additional details about a song.
type SongDetails struct {
	ID          int    `json:"id"`
	SongID      int    `json:"song_id"`
	ReleaseDate string `json:"release_date"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}

// NewSongPayload represents the payload for adding a new song.
type NewSongPayload struct {
	Group string `json:"group" example:"Muse"`
	Song  string `json:"song" example:"Supermassive Black Hole"`
}

// EditSongPayload represents the payload for editing a song.
type EditSongPayload struct {
	Song        Song        `json:"song"`
	SongDetails SongDetails `json:"song_details"`
}
