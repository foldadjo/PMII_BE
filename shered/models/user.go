package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserRole string

const (
	RoleNull   UserRole = "null"
	RoleMapaba UserRole = "mapaba"
	RolePKD    UserRole = "pkd"
	RolePKL    UserRole = "pkl"
	RolePKN    UserRole = "pkn"
)

type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
)

type User struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email            string             `bson:"email" json:"email"`
	PasswordHash     string             `bson:"password_hash" json:"-"`
	FullName         string             `bson:"full_name" json:"full_name"`
	AnggotaID        *string            `bson:"anggota_id,omitempty" json:"anggota_id,omitempty"`
	Role             UserRole           `bson:"role" json:"role"`
	CodeKepengurusan string             `bson:"code_kepengurusan" json:"code_kepengurusan"`
	Gender           Gender             `bson:"gender" json:"gender"`
	BirthDay         time.Time          `bson:"birth_day" json:"birth_day"`
	Active           bool               `bson:"active" json:"active"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time          `bson:"updated_at" json:"updated_at"`
}
