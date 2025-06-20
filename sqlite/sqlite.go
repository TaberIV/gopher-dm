package sqlite

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

type Database struct {
	db *sql.DB
}

var db *sql.DB

func Initialize() (*sql.DB, error) {
	log.Println("Organizing notes...")
	var err error
	db, err = sql.Open("sqlite", "file:campaign.db?cache=shared&mode=rwc")
	if err != nil {
		return nil, err
	}
	// defer db.Close()

	// Create a table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS campaign (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT UNIQUE NOT NULL
	)`)
	if err != nil {
		return nil, err
	}

	// Query data
	// rows, err := db.Query(`SELECT id, title FROM campaign`)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer rows.Close()

	// fmt.Println("Campaigns:")
	// for rows.Next() {
	// 	var id int
	// 	var title string
	// 	err := rows.Scan(&id, &title)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	fmt.Printf("ID: %d, Title: %s\n", id, title)
	// }

	// if err := rows.Err(); err != nil {
	// 	log.Fatal(err)
	// }

	return db, nil
}

func UpdateCampaign(title, refereeID, playerRoleID string) (bool, error) {
	// Insert data
	_, err := db.Exec(`INSERT INTO campaign (title) VALUES (?)`, title)
	if err != nil {
		log.Println(err)
	}
	return false, err
}
