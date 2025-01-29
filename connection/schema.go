package connection

import "fmt"

func (r *Repository) InitSchema() error {

	_, err := r.db.Exec(`
	CREATE TABLE IF NOT EXISTS groups (
	    id SERIAL PRIMARY KEY,
    	name VARCHAR(255) UNIQUE NOT NULL
	);
`)
	if err != nil {
		return fmt.Errorf("error creating groups table: %v", err)
	}

	_, err = r.db.Exec(`
	CREATE TABLE IF NOT EXISTS songs (
	    id SERIAL PRIMARY KEY,
    	name VARCHAR(255) NOT NULL,
    	group_id INTEGER REFERENCES groups(id) ON DELETE CASCADE
	);
`)
	if err != nil {
		return fmt.Errorf("error creating songs table: %v", err)
	}

	_, err = r.db.Exec(`
	CREATE TABLE IF NOT EXISTS song_details (
	    id SERIAL PRIMARY KEY,
		song_id INTEGER REFERENCES songs(id) ON DELETE CASCADE UNIQUE,
    	release_date DATE DEFAULT '1970-01-01',
    	text TEXT DEFAULT 'no information',
    	link VARCHAR(255) DEFAULT 'no information'
	);
`)
	if err != nil {
		return fmt.Errorf("error creating song_details table: %v", err)
	}

	queries := []string{
		`CREATE INDEX IF NOT EXISTS idx_songs_name ON songs (name);`,
		`CREATE INDEX IF NOT EXISTS idx_songs_group_id ON songs (group_id);`,
		`CREATE INDEX IF NOT EXISTS idx_song_details_release_date ON song_details (release_date);`,
	}

	for _, query := range queries {
		_, err := r.db.Exec(query)
		if err != nil {
			return fmt.Errorf("error creating index: %v", err)
		}
	}

	fmt.Println("Schema initialization completed successfully")

	return nil
}
