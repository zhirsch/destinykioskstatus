package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type Manifest struct {
	db    *sql.DB
	stmts map[string]*sql.Stmt
}

func NewManifest(path string) (*Manifest, error) {
	m := &Manifest{stmts: make(map[string]*sql.Stmt)}

	// Open the database connection.
	if sqldb, err := sql.Open("sqlite3", path); err != nil {
		return nil, fmt.Errorf("failed to open manifest db: %v", err)
	} else if sqldb == nil {
		return nil, fmt.Errorf("manifest db is nil")
	} else {
		m.db = sqldb
	}

	// Prepare the SELECT queries.
	if err := m.prepare("DestinyVendorDefinition"); err != nil {
		return nil, err
	}
	if err := m.prepare("DestinyInventoryItemDefinition"); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Manifest) prepare(table string) error {
	stmt, err := m.db.Prepare(fmt.Sprintf("SELECT json FROM %v WHERE id = ?;", table))
	if err != nil {
		return err
	}
	m.stmts[table] = stmt
	return nil
}

func (m *Manifest) get(table string, hash uint32, definition interface{}) {
	var value string
	if err := m.stmts[table].QueryRow(int32(hash)).Scan(&value); err != nil {
		panic(err)
	}
	if err := json.NewDecoder(bytes.NewBuffer([]byte(value))).Decode(definition); err != nil {
		panic(err)
	}
}

type DestinyVendorDefinition struct {
	FailureStrings []string `json:"failureStrings"`
	Summary        struct {
		VendorIdentifier string `json:"vendorIdentifier"`
		VendorName       string `json:"vendorName"`
	} `json:"summary"`
	Hash uint32 `json:"hash"`
}

func (m *Manifest) GetDestinyVendorDefinition(vendorHash uint32) *DestinyVendorDefinition {
	definition := new(DestinyVendorDefinition)
	m.get("DestinyVendorDefinition", vendorHash, definition)
	return definition
}

type DestinyInventoryItemDefinition struct {
	Icon         string   `json:"icon"`
	ItemName     string   `json:"itemName"`
	SourceHashes []uint32 `json:"sourceHashes"`
}

func (m *Manifest) GetDestinyInventoryItemDefinition(itemHash uint32) *DestinyInventoryItemDefinition {
	definition := new(DestinyInventoryItemDefinition)
	m.get("DestinyInventoryItemDefinition", itemHash, definition)
	return definition
}
