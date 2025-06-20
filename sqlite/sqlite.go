package sqlite

import (
	"database/sql"
	"log"

	sq "github.com/Masterminds/squirrel"
	"github.com/Masterminds/structable"
	"github.com/google/uuid"
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
		id INTEGER UNIQUE NOT NULL,
		title TEXT UNIQUE NOT NULL,
		description TEXT DEFAULT '',
		guild_id TEXT NOT NULL,
		referee_id TEXT DEFAULT '',
		player_role_id TEXT DEFAULT '',
		channel_id TEXT DEFAULT '',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT PK_Person PRIMARY KEY (title, guild_id)
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
	Id           uint32 `stbl:"id, UNIQUE NOT NULL"`
	GuildID      string `stbl:"guild_id, NOT NULL"`
	Title        string `stbl:"title, NOT NULL"`
	RefereeID    string `stbl:"referee_id, DEFAULT ''"`
	PlayerRoleID string `stbl:"player_role_id, DEFAULT ''"`
	Description  string `stbl:"description, DEFAULT ''"`
	ChannelID    string `stbl:"channel_id, DEFAULT ''"`
	// createdAt    string
	// updatedAt    string
}

// const uniqueConstraintErrorCode = 2067

const (
	NoChangeCode = iota
	UpdateCode
	InsertCode
)

func FetchUpdateCampaignByTitle(input *Campaign) (obj *Campaign, changeCode int) {
	// ctx, err := db.Begin()
	// if err != nil {
	// 	log.Printf("Error starting transaction: %v", err)
	// 	return
	// }

	// Check if the campaign already exists
	memProxy := sq.NewStmtCacheProxy(db)
	ctx, err := memProxy.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return
	}

	obj = &Campaign{}
	r := structable.New(memProxy, "sqlite").Bind("campaign", obj)
	err = r.LoadWhere(sq.Eq{"title": input.Title}, sq.Eq{"guild_id": input.GuildID})

	switch err {
	case nil:
		log.Printf("Campaign %s already exists with ID %d and GuildID: %s", obj.Title, obj.Id, obj.GuildID)

		var updates bool
		if input.RefereeID != "" {
			obj.RefereeID = input.RefereeID
			updates = true
		}
		if input.PlayerRoleID != "" {
			obj.PlayerRoleID = input.PlayerRoleID
			updates = true
		}
		if input.Description != "" {
			obj.Description = input.Description
			updates = true
		}
		if input.ChannelID != "" {
			obj.ChannelID = input.ChannelID
			updates = true
		}

		if !updates {
			log.Printf("No fields to update for campaign %s", input.Title)
			return
		}

		changeCode = UpdateCode
		r.Update()
	case sql.ErrNoRows:
		log.Printf("Campaign %s does not exist, inserting it...", input.Title)
		obj = input
		obj.Id = uuid.New().ID()
		r.Bind("campaign", obj)

		err = r.Insert()
		if err != nil {
			log.Printf("Error inserting campaign: %v. Rolling back changes... (id is %d)", err, input.Id)
			ctx.Rollback()
			return
		}
		changeCode = InsertCode
		log.Printf("Campaign \"%s\" inserted successfully with ID %d", input.Title, obj.Id)
	default:
		log.Printf("Error checking campaign existence: %v", err)
		return
	}

	return
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
