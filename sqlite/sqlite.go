package sqlite

import (
	"database/sql"
	"log"

	sq "github.com/Masterminds/squirrel"
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
		description TEXT DEFAULT '',
		guild_id TEXT NOT NULL,
		referee_id TEXT DEFAULT '',
		player_role_id TEXT DEFAULT '',
		channel_id TEXT DEFAULT '',
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
	Id           int64
	Title        string
	RefereeID    string
	PlayerRoleID string
	// createdAt    string
	// updatedAt    string
}

// const uniqueConstraintErrorCode = 2067

const (
	NoChangeCode = iota
	UpdateCode
	InsertCode
)

func FetchUpdateCampaignByTitle(input Campaign) (result *Campaign, changeCode int) {
	ctx, err := db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return
	}

	// Check if the campaign already exists
	query, args, err := sq.Select("id").From("campaign").Where(sq.Eq{"title": input.Title}).ToSql()
	if err != nil {
		log.Printf("Error building query: %v", err)
		return
	}

	var id int64

	row := ctx.QueryRow(query, args...)
	err = row.Scan(&id)
	if err == nil {
		log.Printf("Campaign %s already exists with ID %d", input.Title, id)
		setMap := input.makeSetMap()

		if len(setMap) == 1 {
			log.Printf("No fields to update for campaign %s", input.Title)
		} else {
			log.Printf("Updating campaign %s...", input.Title)

			changeCode = UpdateCode
			query, args, err = sq.Update("campaign").SetMap(setMap).Where(sq.Eq{"id": id}).ToSql()
			if err != nil {
				log.Printf("Error building update query: %v", err)
				return
			}
			_, err = ctx.Exec(query, args...)
			if err != nil {
				log.Printf("Error updating campaign: %v", err)
				ctx.Rollback()
				return
			}
		}
	} else if err == sql.ErrNoRows {
		log.Printf("Campaign %s does not exist, inserting it...", input.Title)

		query, args, err = sq.Insert("campaign").SetMap(input.makeSetMap()).ToSql()
		if err != nil {
			log.Printf("Error building insert query: %v", err)
			return
		}

		res, err := ctx.Exec(query, args...)
		if err != nil {
			log.Printf("Error inserting campaign: %v", err)
			ctx.Rollback()
			return
		}

		id, err = res.LastInsertId()
		if err != nil {
			log.Printf("Error getting last insert ID: %v", err)
			ctx.Rollback()
			return
		}
	} else if err != nil {
		log.Printf("Error checking campaign existence: %v", err)
		return
	}

	ctx.Commit()

	result, err = getCampaignById(id)
	return
}

func getCampaignById(id int64) (*Campaign, error) {
	query, args, err := sq.Select("id", "title", "referee_id", "player_role_id").From("campaign").Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		log.Printf("Error building query: %v", err)
		return nil, err
	}
	row := db.QueryRow(query, args...)
	var campaign Campaign
	err = row.Scan(&campaign.Id, &campaign.Title, &campaign.RefereeID, &campaign.PlayerRoleID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No campaign found with ID %d", id)
			return nil, nil
		}
		log.Printf("Error scanning campaign: %v", err)
		return nil, err
	}
	log.Printf("Campaign found: %+v", campaign)
	return &campaign, nil
}

func (input *Campaign) makeSetMap() map[string]interface{} {
	setMap := make(map[string]interface{})

	if input.Title != "" {
		setMap["title"] = input.Title
	}
	if input.RefereeID != "" {
		setMap["referee_id"] = input.RefereeID
	}
	if input.PlayerRoleID != "" {
		setMap["player_role_id"] = input.PlayerRoleID
	}

	return setMap
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

func DeleteCampaign(title string) error {
	_, err := db.Exec(`DELETE FROM campaign WHERE title = ?`, title)
	if err != nil {
		log.Printf("Error deleting campaign %s: %v", title, err)
		return err
	}
	log.Printf("Campaign %s deleted successfully.", title)
	return nil
}
