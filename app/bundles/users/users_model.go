// Package users handles all methods, functions, and structs related to a user
// account
package users

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/krishamoud/game/app/common/db"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2/bson"
)

// User is the basic struct used to hold basic user information
type User struct {
	ID        string  `json:"_id" bson:"_id"`
	Emails    []Email `json:"emails"`
	Username  string  `json:"username"`
	Services  `json:"services"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}

// Email is the basic email field struct for a user
type Email struct {
	Address  string `json:"address"`
	Verified bool   `json:"verified"`
}

// Services handles login information
type Services struct {
	Password `json:"password"`
}

// Password holds the bcrypt-ed password
type Password struct {
	Bcrypt string `json:"bcrypt"`
}

// TODO: change this to something more secret
var tokenEncodeString = "secret"

// NewPassword creates a HMACSHA256 of a given string then returns that SHA
// bcrypt-ed
func NewPassword(password string) (string, error) {
	h := hashPassword(password)
	var err error
	pwHash := fmt.Sprintf("%x", h)
	bcryptStr, err := bcrypt.GenerateFromPassword([]byte(pwHash), 10)
	if err != nil {
		return "", err
	}
	return string(bcryptStr), nil
}

// Authenticate compares a plaintext password with a bcrypt string and returns
// an error if they do not match
func Authenticate(password, bcryptStr string) error {
	h := hashPassword(password)
	var err error
	pwHash := fmt.Sprintf("%x", h)
	err = bcrypt.CompareHashAndPassword([]byte(bcryptStr), []byte(pwHash))
	if err != nil {
		return err
	}
	return nil
}

// showUser finds a single user based on a given id
func showUser(id, reqUserID string) (*User, error) {
	var err error
	u := db.DB.C("users")
	result := &User{}
	err = u.Find(bson.M{"_id": id}).One(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// indexUsers shows all users in the database
func indexUsers() (*[]User, error) {
	var err error
	u := db.DB.C("users")
	var result []User
	err = u.Find(nil).All(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// createUser creates a user with a given email and password
func createUser(email, password string) (*User, error) {
	u := db.DB.C("users")
	id := db.RandomID(17)
	emails := []Email{
		Email{
			Address:  email,
			Verified: false,
		},
	}
	pw, err := NewPassword(password)
	if err != nil {
		return nil, err
	}
	services := Services{
		Password{
			Bcrypt: pw,
		},
	}
	user := &User{
		ID:        id,
		Emails:    emails,
		Username:  email,
		Services:  services,
		CreatedAt: time.Now(),
	}
	err = u.Insert(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// authenticateUser takes an email and password and verifies them agains the
// database.  If successful authenticateUser will return a signed jwt for future
// api calls
func authenticateUser(email, password string) (string, error) {
	var err error
	u := db.DB.C("users")
	result := &User{}
	err = u.Find(bson.M{"emails.address": email}).One(&result)
	if err != nil {
		return "", err
	}
	err = Authenticate(password, result.Services.Password.Bcrypt)
	if err != nil {
		return "", err
	}
	token, err := createToken(result)
	if err != nil {
		return "", err
	}
	return token, nil
}

// hashPassword takes a plaintext password and returns a HMACSHA256 byte slice
// which can be used by the authentication functions
func hashPassword(password string) []byte {
	h := sha256.New()
	h.Write([]byte(password))
	return h.Sum(nil)
}

// createToken generates a jwt and signs it.
func createToken(user *User) (string, error) {
	// create the token
	token := jwt.New(jwt.SigningMethodHS256)
	claims := make(jwt.MapClaims)
	// set some claims
	claims["_id"] = (*user).ID
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
	token.Claims = claims

	//Sign and get the complete encoded token as string
	return token.SignedString([]byte(tokenEncodeString))
}
