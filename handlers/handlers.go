package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/noctusha/music/connection"
	"github.com/noctusha/music/models"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// Handler struct contains the repository for database operations.
type Handler struct {
	Repo *connection.Repository
}

// JSON struct is used for standard JSON responses.
type JSON struct {
	Err   string         `json:"error,omitempty"`
	Songs *[]models.Song `json:"song,omitempty"`
	Text  string         `json:"text,omitempty"`
}

// RespondJSON writes the JSON response with the given status code and payload.
func RespondJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	if payload == nil {
		payload = map[string]string{}
	}

	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, writeErr := w.Write([]byte("Internal Server Error"))
		if writeErr != nil {
			log.Printf("error writing an error in respondJSON %v", writeErr)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)
	_, writeErr := w.Write(response)
	if writeErr != nil {
		log.Printf("error writing response in respondJSON %v", writeErr)
	}
}

// respondJSONError is a helper function to respond with an error message.
func respondJSONError(w http.ResponseWriter, statusCode int, message string) {
	RespondJSON(w, statusCode, JSON{Err: message})
}

// NewHandler creates a new Handler with the given repository.
func NewHandler(repo *connection.Repository) *Handler {
	return &Handler{
		Repo: repo,
	}
}

// ListSongs godoc
// @Summary Get list of songs
// @Description Returns a list of songs with filtering and pagination
// @Tags songs
// @Accept json
// @Produce json
// @Param group query string false "Group name"
// @Param name query string false "Song name"
// @Param releaseDate query string false "Release date"
// @Param text query string false "Song text"
// @Param link query string false "Song link"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} JSON
// @Failure 400 {object} JSON
// @Failure 500 {object} JSON
// @Router /api/songs [get]
// ListSongs handles the request to list songs with optional filters and pagination.
func (h *Handler) ListSongs(w http.ResponseWriter, r *http.Request) {
	var (
		limit       int
		offset      int
		group       string
		name        string
		releaseDate string
		text        string
		link        string
		err         error
	)

	for parameter, vals := range r.URL.Query() {
		switch parameter {
		case "limit":
			limit, err = strconv.Atoi(vals[0])
			if err != nil {
				respondJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid limit format: %v", err))
				return
			}
		case "offset":
			offset, err = strconv.Atoi(vals[0])
			if err != nil {
				respondJSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid offset format: %v", err))
				return
			}
		case "group":
			group = vals[0]
		case "name":
			name = vals[0]
		case "releaseDate":
			releaseDate = vals[0]
		case "text":
			text = vals[0]
		case "link":
			link = vals[0]
		default:
			respondJSONError(w, http.StatusBadRequest, fmt.Sprintf("unrecognized query parameter: %v", parameter))
			return
		}
	}
	songs, err := h.Repo.SongList(group, name, releaseDate, text, link, limit, offset)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to select song from database: %v", err))
		return
	}

	RespondJSON(w, http.StatusOK, JSON{Songs: &songs})
}

// GetText godoc
// @Summary Get song text
// @Description Returns the text of a song with pagination over verses
// @Tags songs
// @Accept json
// @Produce json
// @Param song_id path string true "Song ID"
// @Param page query int false "Page number"
// @Param limit query int false "Number of verses per page"
// @Success 200 {object} JSON
// @Failure 400 {object} JSON
// @Failure 404 {object} JSON
// @Failure 500 {object} JSON
// @Router /api/songs/{song_id}/text [get]
// GetText handles the request to retrieve the text of a song with pagination.
func (h *Handler) GetText(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	songID := vars["song_id"]

	text, ok, err := h.Repo.TextListByID(songID)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to retrieve text: %v", err))
		return
	}

	if !ok {
		respondJSONError(w, http.StatusNotFound, fmt.Sprintf("no such text with song_id: %v", songID))
		return
	}

	verses := strings.Split(text, "\n\n")

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limitParam := r.URL.Query().Get("limit")
	limit, err := strconv.Atoi(limitParam)
	if err != nil || limit <= 0 {
		limit = 1
	}

	start := (page - 1) * limit
	end := start + limit

	if start >= len(verses) {
		respondJSONError(w, http.StatusNotFound, "no more verses")
		return
	}

	if end > len(verses) {
		end = len(verses)
	}

	RespondJSON(w, http.StatusOK, JSON{Text: strings.Join(verses[start:end], "\n\n")})
}

// DeleteSong godoc
// @Summary Delete a song
// @Description Deletes a song by ID
// @Tags songs
// @Accept json
// @Produce json
// @Param song_id path string true "Song ID"
// @Success 200 {object} JSON
// @Failure 400 {object} JSON
// @Failure 500 {object} JSON
// @Router /api/songs/{song_id}/delete [delete]
// DeleteSong handles the request to delete a song.
func (h *Handler) DeleteSong(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	songID := vars["song_id"]

	err := h.Repo.SongDelete(songID)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to delete song: %v", err))
		return
	}

	RespondJSON(w, http.StatusOK, JSON{})
}

// EditSong godoc
// @Summary Edit song data
// @Description Edits song data by ID
// @Tags songs
// @Accept json
// @Produce json
// @Param song_id path string true "Song ID"
// @Param song body models.EditSongPayload true "Song data"
// @Success 200 {object} JSON
// @Failure 400 {object} JSON
// @Failure 500 {object} JSON
// @Router /api/songs/{song_id}/edit [patch]
// EditSong handles the request to edit a song's data.
func (h *Handler) EditSong(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	songID := vars["song_id"]

	song, err := h.Repo.GetSongByID(songID)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to retrieve song: %v", err))
		return
	}

	var payload models.EditSongPayload

	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, fmt.Sprintf("failed to decode updatedSong body: %v", err))
		return
	}

	if payload.Song.Name != "" {
		song.Name = payload.Song.Name
	}
	if payload.Song.GroupID != 0 {
		song.GroupID = payload.Song.GroupID
	}

	songDetails, err := h.Repo.GetSongDetailsByID(songID)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to retrieve song: %v", err))
		return
	}

	if payload.SongDetails.ReleaseDate != "" {
		songDetails.ReleaseDate = payload.SongDetails.ReleaseDate
	}
	if payload.SongDetails.Text != "" {
		songDetails.Text = payload.SongDetails.Text
	}
	if payload.SongDetails.Link != "" {
		songDetails.Link = payload.SongDetails.Link
	}

	err = h.Repo.UpdateSong(song, songDetails)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to update song: %v", err))
		return
	}

	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"song":         song,
		"song_details": songDetails,
	})
}

// NewSong godoc
// @Summary Add a new song
// @Description Adds a new song and saves it to the database
// @Tags songs
// @Accept json
// @Produce json
// @Param song body models.NewSongPayload true "New song"
// @Success 201 {object} models.Song
// @Failure 400 {object} JSON
// @Failure 500 {object} JSON
// @Router /api/songs/new [post]
// NewSong handles the request to add a new song.
func (h *Handler) NewSong(w http.ResponseWriter, r *http.Request) {
	var payload models.NewSongPayload

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, fmt.Sprintf("failed to decode song: %v", err))
		return
	}

	if payload.Group == "" {
		respondJSONError(w, http.StatusBadRequest, fmt.Sprintf("no group name"))
		return
	}
	if payload.Song == "" {
		respondJSONError(w, http.StatusBadRequest, fmt.Sprintf("no song name"))
		return
	}

	apiBaseURL := os.Getenv("EXTERNAL_API_URL")
	if apiBaseURL == "" {
		respondJSONError(w, http.StatusInternalServerError, "External API URL is not configured")
		return
	}

	apiURL := fmt.Sprintf("%s/info?group=%s&song=%s", apiBaseURL, url.QueryEscape(payload.Group), url.QueryEscape(payload.Song))
	resp, err := http.Get(apiURL)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, fmt.Sprintf("Error making request to external API: %v", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respondJSONError(w, http.StatusInternalServerError, fmt.Sprintf("External API returned status %d", resp.StatusCode))
		return
	}

	var details models.SongDetails
	err = json.NewDecoder(resp.Body).Decode(&details)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, fmt.Sprintf("Error parsing external API response: %v", err))
		return
	}

	groupID, err := h.Repo.GetGroupID(payload.Group)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to retrieve groupID: %v", err))
		return
	}

	if groupID == 0 {
		groupID, err = h.Repo.NewGroup(payload.Group)
		if err != nil {
			respondJSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to retrieve groupID: %v", err))
			return
		}
	}

	song := models.Song{Name: payload.Song, GroupID: groupID}

	err = h.Repo.CreateSongWithDetails(song, details)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, fmt.Sprintf("failed to create song: %v", err))
		return
	}

	RespondJSON(w, http.StatusCreated, song)
}
