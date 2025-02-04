// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0

package db

import (
	"database/sql"
)

type Registry struct {
	ID         int64
	Structure  string
	Controller string
	Prefix     string
	CreatedAt  int64
	ChainID    int64
}

type SqliteMaster struct {
	Name string
	Sql  string
}

type SystemAcl struct {
	TableID    int64
	Controller string
	Privileges int
	ChainID    int64
	CreatedAt  int64
	UpdatedAt  sql.NullInt64
}

type SystemController struct {
	ChainID    int64
	TableID    int64
	Controller string
}

type SystemEvmBlock struct {
	ChainID     int64
	BlockNumber int64
	Timestamp   int64
}

type SystemEvmEvent struct {
	ChainID     int64
	EventJson   string
	EventType   string
	Address     string
	Topics      string
	Data        []byte
	BlockNumber int64
	TxHash      string
	TxIndex     uint
	BlockHash   string
	EventIndex  uint
}

type SystemID struct {
	ID string
}

type SystemPendingTx struct {
	ChainID        int64
	Address        string
	Hash           string
	Nonce          int64
	BumpPriceCount int
	CreatedAt      int64
	UpdatedAt      sql.NullInt64
}

type SystemTxnProcessor struct {
	ChainID     int64
	BlockNumber int64
}

type SystemTxnReceipt struct {
	ChainID       int64
	BlockNumber   int64
	IndexInBlock  int64
	TxnHash       string
	Error         sql.NullString
	TableID       sql.NullInt64
	ErrorEventIdx sql.NullInt64
}
