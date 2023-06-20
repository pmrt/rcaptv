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

var Vods = newVodsTable("public", "vods", "")

type vodsTable struct {
	postgres.Table

	// Columns
	VideoID         postgres.ColumnString
	StreamID        postgres.ColumnString
	BcID            postgres.ColumnString
	CreatedAt       postgres.ColumnTimestamp
	PublishedAt     postgres.ColumnTimestamp
	DurationSeconds postgres.ColumnInteger
	Lang            postgres.ColumnString
	ThumbnailURL    postgres.ColumnString
	Title           postgres.ColumnString
	ViewCount       postgres.ColumnInteger

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type VodsTable struct {
	vodsTable

	EXCLUDED vodsTable
}

// AS creates new VodsTable with assigned alias
func (a VodsTable) AS(alias string) *VodsTable {
	return newVodsTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new VodsTable with assigned schema name
func (a VodsTable) FromSchema(schemaName string) *VodsTable {
	return newVodsTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new VodsTable with assigned table prefix
func (a VodsTable) WithPrefix(prefix string) *VodsTable {
	return newVodsTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new VodsTable with assigned table suffix
func (a VodsTable) WithSuffix(suffix string) *VodsTable {
	return newVodsTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newVodsTable(schemaName, tableName, alias string) *VodsTable {
	return &VodsTable{
		vodsTable: newVodsTableImpl(schemaName, tableName, alias),
		EXCLUDED:  newVodsTableImpl("", "excluded", ""),
	}
}

func newVodsTableImpl(schemaName, tableName, alias string) vodsTable {
	var (
		VideoIDColumn         = postgres.StringColumn("video_id")
		StreamIDColumn        = postgres.StringColumn("stream_id")
		BcIDColumn            = postgres.StringColumn("bc_id")
		CreatedAtColumn       = postgres.TimestampColumn("created_at")
		PublishedAtColumn     = postgres.TimestampColumn("published_at")
		DurationSecondsColumn = postgres.IntegerColumn("duration_seconds")
		LangColumn            = postgres.StringColumn("lang")
		ThumbnailURLColumn    = postgres.StringColumn("thumbnail_url")
		TitleColumn           = postgres.StringColumn("title")
		ViewCountColumn       = postgres.IntegerColumn("view_count")
		allColumns            = postgres.ColumnList{VideoIDColumn, StreamIDColumn, BcIDColumn, CreatedAtColumn, PublishedAtColumn, DurationSecondsColumn, LangColumn, ThumbnailURLColumn, TitleColumn, ViewCountColumn}
		mutableColumns        = postgres.ColumnList{StreamIDColumn, BcIDColumn, CreatedAtColumn, PublishedAtColumn, DurationSecondsColumn, LangColumn, ThumbnailURLColumn, TitleColumn, ViewCountColumn}
	)

	return vodsTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		VideoID:         VideoIDColumn,
		StreamID:        StreamIDColumn,
		BcID:            BcIDColumn,
		CreatedAt:       CreatedAtColumn,
		PublishedAt:     PublishedAtColumn,
		DurationSeconds: DurationSecondsColumn,
		Lang:            LangColumn,
		ThumbnailURL:    ThumbnailURLColumn,
		Title:           TitleColumn,
		ViewCount:       ViewCountColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}