package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
	"github.com/google/uuid"
)

func main() {
	db, err := sql.Open("sqlite", "photos.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 1. Get the source photo (the latest one)
	var id, filepathSrc, thumbPathSrc, notes, exif string
	var lat, lon float64
	
	row := db.QueryRow("SELECT id, filepath, thumbnail_path, lat, lon, notes, exif_data FROM photos ORDER BY day DESC LIMIT 1")
	err = row.Scan(&id, &filepathSrc, &thumbPathSrc, &lat, &lon, &notes, &exif)
	if err != nil {
		log.Fatalf("No photos found to seed from: %v", err)
	}

	fmt.Printf("Seeding from photo ID: %s (%s)\n", id, filepathSrc)

	// 2. Define Range: Nov 15 2025 to Dec 29 2025 (Assuming current year 2025 from user context, logic works for any recent year)
    // Actually, let's just go back 45 days from today.
    
    end := time.Now()
    start := end.AddDate(0, 0, -45) 
    
    // Create 'uploads' dir if safe? It exists.
    
    count := 0
    for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
        dayStr := d.Format("2006-01-02")
        
        // Skip if exists
        var exists bool
        db.QueryRow("SELECT EXISTS(SELECT 1 FROM photos WHERE day = ?)", dayStr).Scan(&exists)
        if exists {
            fmt.Printf("Skipping %s (exists)\n", dayStr)
            continue
        }

        // Generate new ID
        newID := uuid.New().String()
        
        // Copy File
        newFilename := fmt.Sprintf("%s.jpg", newID)
        newPath := filepath.Join("uploads", newFilename)
        if err := copyFile(filepathSrc[1:], newPath); err != nil { // Remove leading '/' from filepathSrc if present?
             // filepathSrc in DB is likely "/uploads/..." from my previous code?
             // Actually my code saves as "/uploads/..." (abs-ish path string for URL) but on disk it is relative to cwd "uploads/..."
             // Let's check filepathSrc. If it starts with "/", strip it for disk access.
             src := filepathSrc
             if src[0] == '/' { src = src[1:] }
             copyFile(src, newPath)
        }
        
        // Copy Thumbnail
        newThumbFilename := fmt.Sprintf("%s_thumb.jpg", newID)
        newThumbPath := filepath.Join("uploads", newThumbFilename)
         // src thumb
         srcThumb := thumbPathSrc
         if srcThumb[0] == '/' { srcThumb = srcThumb[1:] }
         copyFile(srcThumb, newThumbPath)
         
         // Insert DB
         _, err = db.Exec(`INSERT INTO photos (day, id, filepath, thumbnail_path, lat, lon, notes, exif_data, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
            dayStr, newID, "/"+newPath, "/"+newThumbPath, lat, lon, fmt.Sprintf("Seeded clone %s", dayStr), exif, time.Now())
        if err != nil {
            log.Printf("Failed to insert %s: %v", dayStr, err)
        } else {
            count++
        }
    }
    
    fmt.Printf("Seeded %d new photos.\n", count)
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
