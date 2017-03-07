package db

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type (
	tableEnum string
	stmtEnum  string
)

const (
	tableBungieUsers       tableEnum = "BungieUsers"
	tableDestinyUsers      tableEnum = "DestinyUsers"
	tableDestinyCharacters tableEnum = "DestinyCharacters"

	stmtCreate stmtEnum = "CREATE"
	stmtInsert stmtEnum = "INSERT"
	stmtSelect stmtEnum = "SELECT"
)

var (
	tableStmtSQL = map[tableEnum]map[stmtEnum]string{
		tableBungieUsers: {
			stmtCreate: `
CREATE TABLE IF NOT EXISTS BungieUsers(
    MembershipID      TEXT PRIMARY KEY,
    DisplayName       TEXT,
    TokenAccessToken  TEXT,
    TokenRefreshToken TEXT,
    TokenExpiry       DATETIME
);
`,
			stmtInsert: `
INSERT OR REPLACE INTO BungieUsers(
    MembershipID,
    DisplayName,
    TokenAccessToken,
    TokenRefreshToken,
    TokenExpiry
) VALUES(?, ?, ?, ?, ?);
`,
			stmtSelect: `
SELECT
    DisplayName,
    TokenAccessToken,
    TokenRefreshToken,
    TokenExpiry
FROM
    BungieUsers
WHERE
    MembershipID = ?;
`,
		},

		tableDestinyUsers: {
			stmtCreate: `
CREATE TABLE IF NOT EXISTS DestinyUsers(
    MembershipType     INT64,
    MembershipID       TEXT,
    DisplayName        TEXT,
    BungieMembershipID TEXT,
    PRIMARY KEY (MembershipType, MembershipID)
);
CREATE INDEX IF NOT EXISTS DestinyUsers_BungieMembershipID
ON DestinyUsers (BungieMembershipID);
`,
			stmtInsert: `
INSERT OR REPLACE INTO DestinyUsers(
    MembershipType,
    MembershipID,
    DisplayName,
    BungieMembershipID
) VALUES(?, ?, ?, ?);
`,
			stmtSelect: `
SELECT
    MembershipType,
    MembershipID,
    DisplayName
FROM
    DestinyUsers
WHERE
    BungieMembershipID = ?;
`,
		},

		tableDestinyCharacters: {
			stmtCreate: `
CREATE TABLE IF NOT EXISTS DestinyCharacters(
    CharacterID           TEXT PRIMARY KEY,
    ClassName             TEXT,
    DestinyMembershipType INT64,
    DestinyMembershipID   TEXT
);
CREATE INDEX IF NOT EXISTS DestinyCharacters_DestinyMembershipType_DestinyMembershipID
ON DestinyCharacters (DestinyMembershipType, DestinyMembershipID);
`,
			stmtInsert: `
INSERT OR REPLACE INTO DestinyCharacters(
    CharacterID,
    ClassName,
    DestinyMembershipType,
    DestinyMembershipID
) VALUES(?, ?, ?, ?);
`,
			stmtSelect: `
SELECT
    CharacterID,
    ClassName
FROM
    DestinyCharacters
WHERE
    DestinyMembershipType = ? AND
    DestinyMembershipID = ?;
`,
		},
	}
)

type table struct {
	stmts map[stmtEnum]*sql.Stmt
}

type DB struct {
	db *sql.DB

	tables map[tableEnum]*table
}

func NewDB(path string) (*DB, error) {
	db := &DB{
		tables: make(map[tableEnum]*table),
	}

	// Create the database connection.
	if sqldb, err := sql.Open("sqlite3", path); err != nil {
		return nil, fmt.Errorf("failed to open db: %v", err)
	} else if sqldb == nil {
		return nil, fmt.Errorf("db is nil")
	} else {
		db.db = sqldb
	}

	for tbl, stmts := range tableStmtSQL {
		db.tables[tbl] = &table{
			stmts: make(map[stmtEnum]*sql.Stmt),
		}

		// Create the table.
		if _, err := db.db.Exec(stmts[stmtCreate]); err != nil {
			return nil, fmt.Errorf("failed to create table %v: %v", tbl, err)
		}

		// Prepare all the statements.
		for stmt, sql := range stmts {
			if stmt == stmtCreate {
				continue
			}
			if s, err := db.db.Prepare(sql); err != nil {
				return nil, fmt.Errorf("failed to prepare %v on table %v: %v", stmt, tbl, err)
			} else {
				db.tables[tbl].stmts[stmt] = s
			}
		}
	}

	return db, nil
}
