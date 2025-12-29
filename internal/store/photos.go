package store

import (
	"database/sql"
    "time"
)

type Photo struct {
	Day           string // YYYY-MM-DD
	ID            string
	Filepath      string
	ThumbnailPath string
	Lat           float64
	Lon           float64
	Notes         string
    ExifData      string
	CreatedAt     time.Time
}

type PhotoStore struct {
	db *sql.DB
}

func NewPhotoStore(db *sql.DB) *PhotoStore {
	return &PhotoStore{db: db}
}

func (s *PhotoStore) Save(p *Photo) error {
	query := `
    INSERT INTO photos (day, id, filepath, thumbnail_path, lat, lon, notes, exif_data, created_at)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    ON CONFLICT(day) DO UPDATE SET
        id=excluded.id,
        filepath=excluded.filepath,
        thumbnail_path=excluded.thumbnail_path,
        lat=excluded.lat,
        lon=excluded.lon,
        notes=excluded.notes,
        exif_data=excluded.exif_data,
        created_at=excluded.created_at;
    `
    _, err := s.db.Exec(query, p.Day, p.ID, p.Filepath, p.ThumbnailPath, p.Lat, p.Lon, p.Notes, p.ExifData, p.CreatedAt)
    return err
}

func (s *PhotoStore) GetByDay(day string) (*Photo, error) {
    p := &Photo{}
    err := s.db.QueryRow("SELECT day, id, filepath, thumbnail_path, lat, lon, notes, exif_data, created_at FROM photos WHERE day = ?", day).Scan(
        &p.Day, &p.ID, &p.Filepath, &p.ThumbnailPath, &p.Lat, &p.Lon, &p.Notes, &p.ExifData, &p.CreatedAt,
    )
    if err != nil {
        return nil, err
    }
    return p, nil
}

func (s *PhotoStore) List(limit int) ([]Photo, error) {
    rows, err := s.db.Query("SELECT day, id, filepath, thumbnail_path, lat, lon, notes, exif_data, created_at FROM photos ORDER BY day DESC LIMIT ?", limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    photos := []Photo{}
    for rows.Next() {
        var p Photo
        if err := rows.Scan(&p.Day, &p.ID, &p.Filepath, &p.ThumbnailPath, &p.Lat, &p.Lon, &p.Notes, &p.ExifData, &p.CreatedAt); err != nil {
            return nil, err
        }
        photos = append(photos, p)
    }
    return photos, nil
}
