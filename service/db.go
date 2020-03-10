package service

import (
	"database/sql"
	"fmt"

	"github.com/ekr-paolo-carraro/go-jwt/domain"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

//PostgresService acts as service for db
type PostgresService struct {
	DB *sql.DB
}

//InitDBService starts postgresql db connection and return err in case of failure
func InitDBService() (*PostgresService, error) {

	connection := "host=localhost port=5432 user=srvuser password=ekr dbname=test sslmode=disable"
	db, err := sql.Open("postgres", connection)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &PostgresService{db}, nil
}

//AddUser add a user before check the existence
func (ps *PostgresService) AddUser(user domain.User) (int, error) {

	candidate, err := ps.GetUser(user.Email)
	if err != nil {
		return 0, fmt.Errorf("Error on checking user existance: %v", err.Error())
	}

	if candidate != nil {
		return 0, fmt.Errorf("User already exists. Change username.")
	}

	var sqlInsert string = "INSERT INTO users (email,password) VALUES ($1,$2) RETURNING id;"
	encryptPws, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
	if err != nil {
		return 0, fmt.Errorf("Error on crypt password: %v", err.Error())
	}

	err = ps.DB.QueryRow(sqlInsert, user.Email, string(encryptPws)).Scan(&user.ID)
	if err != nil {
		return 0, fmt.Errorf("Error on insert new user: %v", err.Error())
	}

	if user.ID == 0 {
		return 0, fmt.Errorf("Error on insert new user: %v", err.Error())
	}

	return user.ID, nil
}

//GetUser search user in db by username/email
func (ps *PostgresService) GetUser(username string) (*domain.User, error) {
	var sqlCheckExistence string = "SELECT * FROM users WHERE email = $1;"
	qr, err := ps.DB.Query(sqlCheckExistence, username)
	if err != nil {
		return nil, fmt.Errorf("error checking user existence: %v", err.Error())
	}

	foundUser := domain.User{}
	for qr.Next() {
		qr.Scan(&foundUser.ID, &foundUser.Email, &foundUser.Password)
	}

	if foundUser.ID == 0 && foundUser.Email == "" {
		return nil, nil
	}

	return &foundUser, nil
}
