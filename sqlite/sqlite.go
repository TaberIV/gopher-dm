package sqlite

import (
	"database/sql"
	"errors"
	"log"
	"strings"

	"modernc.org/sqlite"
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
		title TEXT UNIQUE NOT NULL,
		referee_id TEXT DEFAULT '',
		player_role_id TEXT DEFAULT '',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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

type Campaign struct {
	Id           int
	Title        string
	RefereeID    string
	PlayerRoleID string
	// createdAt    string
	// updatedAt    string
}

const uniqueConstraintErrorCode = 2067

func UpdateCampaign(input Campaign) (Campaign, error) {
	campaign := Campaign{
		Title:        input.Title,
		RefereeID:    input.RefereeID,
		PlayerRoleID: input.PlayerRoleID,
	}

	// Insert data
	_, err := db.Exec(`INSERT INTO campaign (title) VALUES (?)`, input.Title)
	var e *sqlite.Error
	if err != nil {
		if errors.As(err, &e); e.Code() == uniqueConstraintErrorCode {
			log.Printf("Campaign %s already exists. Updating...", input.Title)

			// Build update query dynamically based on provided fields
			setClauses := []string{}
			args := []any{}

			if input.RefereeID != "" {
				setClauses = append(setClauses, "referee_id = ?")
				args = append(args, input.RefereeID)
			}
			if input.PlayerRoleID != "" {
				setClauses = append(setClauses, "player_role_id = ?")
				args = append(args, input.PlayerRoleID)
			}

			if len(setClauses) == 0 {
				log.Printf("No fields to update for campaign %s.", input.Title)
			} else {
				query := "UPDATE campaign SET " +
					strings.Join(setClauses, ", ") +
					" WHERE title = ?"
				args = append(args, input.Title)

				_, err = db.Exec(query, args...)

				if err != nil {
					log.Printf("Error updating campaign %s: %v", input.Title, err)
					return campaign, err
				}
				log.Printf("Campaign %s updated successfully.", input.Title)
			}
		} else {
			log.Fatal(err)
		}
	}

	// Get the full campaign details from the database
	row := db.QueryRow(`SELECT id, title, referee_id, player_role_id FROM campaign WHERE title = ?`, input.Title)
	err = row.Scan(&campaign.Id, &campaign.Title, &campaign.RefereeID, &campaign.PlayerRoleID)
	if err != nil {
		log.Printf("Error retrieving campaign %s: %v", input.Title, err)
	}
	log.Printf("Campaign %s created successfully.", input.Title)

	return campaign, err
}

func GetAllCampaigns() ([]Campaign, error) {
	rows, err := db.Query(`SELECT id, title, referee_id, player_role_id FROM campaign`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var campaigns []Campaign
	for rows.Next() {
		var campaign Campaign
		err := rows.Scan(&campaign.Id, &campaign.Title, &campaign.RefereeID, &campaign.PlayerRoleID)
		if err != nil {
			log.Fatal(err)
		}
		campaigns = append(campaigns, campaign)
	}

	return campaigns, nil
}
