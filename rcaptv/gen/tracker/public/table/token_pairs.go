//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package table

import (
	"github.com/go-jet/jet/v2/postgres"
)

var TokenPairs = newTokenPairsTable("public", "token_pairs", "")

type tokenPairsTable struct {
	postgres.Table

	// Columns
	TokenPairID    postgres.ColumnInteger
	UserID         postgres.ColumnInteger
	AccessToken    postgres.ColumnString
	RefreshToken   postgres.ColumnString
	ExpiresAt      postgres.ColumnTimestamp
	LastModifiedAt postgres.ColumnTimestamp

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type TokenPairsTable struct {
	tokenPairsTable

	EXCLUDED tokenPairsTable
}

// AS creates new TokenPairsTable with assigned alias
func (a TokenPairsTable) AS(alias string) *TokenPairsTable {
	return newTokenPairsTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new TokenPairsTable with assigned schema name
func (a TokenPairsTable) FromSchema(schemaName string) *TokenPairsTable {
	return newTokenPairsTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new TokenPairsTable with assigned table prefix
func (a TokenPairsTable) WithPrefix(prefix string) *TokenPairsTable {
	return newTokenPairsTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new TokenPairsTable with assigned table suffix
func (a TokenPairsTable) WithSuffix(suffix string) *TokenPairsTable {
	return newTokenPairsTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newTokenPairsTable(schemaName, tableName, alias string) *TokenPairsTable {
	return &TokenPairsTable{
		tokenPairsTable: newTokenPairsTableImpl(schemaName, tableName, alias),
		EXCLUDED:        newTokenPairsTableImpl("", "excluded", ""),
	}
}

func newTokenPairsTableImpl(schemaName, tableName, alias string) tokenPairsTable {
	var (
		TokenPairIDColumn    = postgres.IntegerColumn("token_pair_id")
		UserIDColumn         = postgres.IntegerColumn("user_id")
		AccessTokenColumn    = postgres.StringColumn("access_token")
		RefreshTokenColumn   = postgres.StringColumn("refresh_token")
		ExpiresAtColumn      = postgres.TimestampColumn("expires_at")
		LastModifiedAtColumn = postgres.TimestampColumn("last_modified_at")
		allColumns           = postgres.ColumnList{TokenPairIDColumn, UserIDColumn, AccessTokenColumn, RefreshTokenColumn, ExpiresAtColumn, LastModifiedAtColumn}
		mutableColumns       = postgres.ColumnList{UserIDColumn, AccessTokenColumn, RefreshTokenColumn, ExpiresAtColumn, LastModifiedAtColumn}
	)

	return tokenPairsTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		TokenPairID:    TokenPairIDColumn,
		UserID:         UserIDColumn,
		AccessToken:    AccessTokenColumn,
		RefreshToken:   RefreshTokenColumn,
		ExpiresAt:      ExpiresAtColumn,
		LastModifiedAt: LastModifiedAtColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}
