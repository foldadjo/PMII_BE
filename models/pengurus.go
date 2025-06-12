package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PengurusLevel string

const (
	LevelPB  PengurusLevel = "PB"
	LevelPKC PengurusLevel = "PKC"
	LevelPC  PengurusLevel = "PC"
	LevelPK  PengurusLevel = "PK"
	LevelPR  PengurusLevel = "PR"
)

type Pengurus struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID        primitive.ObjectID `bson:"user_id" json:"user_id"`
	Level         PengurusLevel      `bson:"level" json:"level"`
	Wilayah       *string            `bson:"wilayah,omitempty" json:"wilayah,omitempty"`
	Cabang        *string            `bson:"cabang,omitempty" json:"cabang,omitempty"`
	Komisariat    *string            `bson:"komisariat,omitempty" json:"komisariat,omitempty"`
	Jabatan       string             `bson:"jabatan" json:"jabatan"`
	Aktif         bool               `bson:"aktif" json:"aktif"`
	MulaiJabatan  time.Time          `bson:"mulai_jabatan" json:"mulai_jabatan"`
	AkhirJabatan  time.Time          `bson:"akhir_jabatan" json:"akhir_jabatan"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
} 