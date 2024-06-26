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

var Clips = newClipsTable("public", "clips", "")

type clipsTable struct {
	postgres.Table

	// Columns
	ClipID          postgres.ColumnString
	BcID            postgres.ColumnString
	VideoID         postgres.ColumnString
	CreatedAt       postgres.ColumnTimestamp
	CreatorID       postgres.ColumnString
	CreatorName     postgres.ColumnString
	Title           postgres.ColumnString
	GameID          postgres.ColumnString
	Lang            postgres.ColumnString
	ThumbnailURL    postgres.ColumnString
	DurationSeconds postgres.ColumnFloat
	ViewCount       postgres.ColumnInteger
	VodOffset       postgres.ColumnInteger

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type ClipsTable struct {
	clipsTable

	EXCLUDED clipsTable
}

// AS creates new ClipsTable with assigned alias
func (a ClipsTable) AS(alias string) *ClipsTable {
	return newClipsTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new ClipsTable with assigned schema name
func (a ClipsTable) FromSchema(schemaName string) *ClipsTable {
	return newClipsTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new ClipsTable with assigned table prefix
func (a ClipsTable) WithPrefix(prefix string) *ClipsTable {
	return newClipsTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new ClipsTable with assigned table suffix
func (a ClipsTable) WithSuffix(suffix string) *ClipsTable {
	return newClipsTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newClipsTable(schemaName, tableName, alias string) *ClipsTable {
	return &ClipsTable{
		clipsTable: newClipsTableImpl(schemaName, tableName, alias),
		EXCLUDED:   newClipsTableImpl("", "excluded", ""),
	}
}

func newClipsTableImpl(schemaName, tableName, alias string) clipsTable {
	var (
		ClipIDColumn          = postgres.StringColumn("clip_id")
		BcIDColumn            = postgres.StringColumn("bc_id")
		VideoIDColumn         = postgres.StringColumn("video_id")
		CreatedAtColumn       = postgres.TimestampColumn("created_at")
		CreatorIDColumn       = postgres.StringColumn("creator_id")
		CreatorNameColumn     = postgres.StringColumn("creator_name")
		TitleColumn           = postgres.StringColumn("title")
		GameIDColumn          = postgres.StringColumn("game_id")
		LangColumn            = postgres.StringColumn("lang")
		ThumbnailURLColumn    = postgres.StringColumn("thumbnail_url")
		DurationSecondsColumn = postgres.FloatColumn("duration_seconds")
		ViewCountColumn       = postgres.IntegerColumn("view_count")
		VodOffsetColumn       = postgres.IntegerColumn("vod_offset")
		allColumns            = postgres.ColumnList{ClipIDColumn, BcIDColumn, VideoIDColumn, CreatedAtColumn, CreatorIDColumn, CreatorNameColumn, TitleColumn, GameIDColumn, LangColumn, ThumbnailURLColumn, DurationSecondsColumn, ViewCountColumn, VodOffsetColumn}
		mutableColumns        = postgres.ColumnList{BcIDColumn, VideoIDColumn, CreatedAtColumn, CreatorIDColumn, CreatorNameColumn, TitleColumn, GameIDColumn, LangColumn, ThumbnailURLColumn, DurationSecondsColumn, ViewCountColumn, VodOffsetColumn}
	)

	return clipsTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		ClipID:          ClipIDColumn,
		BcID:            BcIDColumn,
		VideoID:         VideoIDColumn,
		CreatedAt:       CreatedAtColumn,
		CreatorID:       CreatorIDColumn,
		CreatorName:     CreatorNameColumn,
		Title:           TitleColumn,
		GameID:          GameIDColumn,
		Lang:            LangColumn,
		ThumbnailURL:    ThumbnailURLColumn,
		DurationSeconds: DurationSecondsColumn,
		ViewCount:       ViewCountColumn,
		VodOffset:       VodOffsetColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}
