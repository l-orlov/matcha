package models

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type (
	UserToCreate struct {
		Email     string `json:"email" binding:"required"`
		Username  string `json:"username" binding:"required"`
		FirstName string `json:"firstName" binding:"required"`
		LastName  string `json:"lastName" binding:"required"`
		Password  string `json:"password" binding:"required"`
	}
	UserToSignIn struct {
		Username    string `json:"username" binding:"required"`
		Password    string `json:"password" binding:"required"`
		Fingerprint string `json:"fingerprint" binding:"required"`
	}
	User struct {
		ID               uint64 `json:"id" binding:"required" db:"id"`
		Email            string `json:"email" binding:"required" db:"email"`
		Username         string `json:"username" binding:"required" db:"username"`
		FirstName        string `json:"firstName" binding:"required" db:"first_name"`
		LastName         string `json:"lastName" binding:"required" db:"last_name"`
		Password         string `json:"-" db:"password"`
		IsEmailConfirmed bool   `json:"isEmailConfirmed" db:"is_email_confirmed"`
	}
	UserPassword struct {
		ID       uint64 `json:"id" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	UserPasswordToChange struct {
		ID          uint64 `json:"id" binding:"required"`
		OldPassword string `json:"oldPassword" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required"`
	}
	UserProfile struct {
		ID                uint64        `json:"id" binding:"required"`
		Email             string        `json:"email" binding:"required"`
		Username          string        `json:"username" binding:"required"`
		FirstName         string        `json:"firstName" binding:"required"`
		LastName          string        `json:"lastName" binding:"required"`
		IsEmailConfirmed  bool          `json:"isEmailConfirmed"`
		Gender            int           `json:"gender"`
		SexualPreferences int           `json:"sexualPreferences"`
		Biography         string        `json:"biography"`
		Tags              []string      `json:"tags"`
		AvatarPath        string        `json:"avatarPath"`
		AvatarURL         string        `json:"avatarURL"`
		Pictures          []UserPicture `json:"pictures"`
		LikesNum          int           `json:"likesNum"`
		ViewsNum          int           `json:"viewsNum"`
		GPSPosition       string        `json:"gpsPosition"`
	}
	UserPicture struct {
		UUID        uuid.UUID `json:"uuid" db:"uuid"`
		UserID      uint64    `json:"userId" db:"user_id"`
		PicturePath string    `json:"picturePath" db:"picture_path"`
		PictureURL  string    `json:"pictureURL"`
	}
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
