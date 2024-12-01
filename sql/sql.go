package sql

import (
	"database/sql"
	"fmt"
	"github/lnj/inventory/model"
	"sync"

	"github.com/rs/zerolog/log"

	_ "github.com/mattn/go-sqlite3"
)

var sqlock sync.Mutex

type Store struct {
	DB   *sql.DB
	Type string
	Path string
}

func NewStoreSQLite(path string) *Store {
	return &Store{
		Type: "sqlite",
		Path: path,
	}
}

func (s *Store) Open() {
	fmt.Println("Open db")
	db, err := sql.Open("sqlite3", s.Path)
	if err != nil {
		log.Panic().Msgf("Open sqlite error - '%s'", err.Error())
	}
	s.DB = db
	log.Info().Msg("The sqlite db is open.")

	s.createTables()
}

func (s *Store) createTables() {
	_, err := s.DB.Exec("create table IF NOT EXISTS USERS (id INTEGER PRIMARY KEY, UserId TEXT UNIQUE, tKey TEXT UNIQUE, Password TEXT CHECK(Password <> ''));")
	if err != nil {
		log.Panic().Msgf("Create USERS Table error - '%s'", err.Error())
	}
	_, err = s.DB.Exec("create table IF NOT EXISTS ITEMS (id INTEGER PRIMARY KEY, tKey TEXT, dKey TEXT UNIQUE, name TEXT CHECK(name <> ''), description TEXT, value TEXT, purchasedate TEXT, serialnum TEXT);")
	if err != nil {
		log.Panic().Msgf("Create ITEMS Table error - '%s'", err.Error())
	}
}

func (s *Store) InsertUser(userId, tKey, password string) {
	sqlock.Lock()
	defer sqlock.Unlock()

	_, err := s.DB.Exec(fmt.Sprintf("insert into USERS(UserId, tKey, Password) values('%s', '%s', '%s');", userId, tKey, password))
	if err != nil {
		log.Panic().Msgf("insert into USERS table error - '%s'", err.Error())
	}
}

func (s *Store) InsertItem(tKey, dKey, name, description, value, purchasedate, serialnum string) {
	sqlock.Lock()
	defer sqlock.Unlock()

	_, err := s.DB.Exec(fmt.Sprintf("insert into ITEMS(tKey, dKey, name, description, value, purchasedate, serialnum) values('%s', '%s', '%s', '%s', '%s', '%s', '%s');", tKey, dKey, name, description, value, purchasedate, serialnum))
	if err != nil {
		log.Panic().Msgf("insert into ITEMS table error - '%s'", err.Error())
	}
}

func (s *Store) UpdateItem(dKey, name, description, value, purchasedate, serialnum string) {
	sqlock.Lock()
	defer sqlock.Unlock()

	_, err := s.DB.Exec(fmt.Sprintf("update ITEMS set name='%s', description='%s', value='%s', purchasedate='%s', serialnum='%s' where dKey='%s';", name, description, value, purchasedate, serialnum, dKey))
	if err != nil {
		log.Panic().Msgf("update USERS table error - '%s'", err.Error())
	}
}

func (s *Store) SelectAllUsers() []string {
	sqlock.Lock()
	defer sqlock.Unlock()

	var userList []string

	rows, err := s.DB.Query("select userId from USERS")
	if err != nil {
		log.Info().Msgf("select query failed. '%s'", err.Error())
		return userList
	}
	for rows.Next() {
		var user string
		rows.Scan(&user)
		userList = append(userList, user)
	}
	fmt.Println("selectAllUser:", userList)
	return userList
}

func (s *Store) SelectUser(userId string) map[string]string {
	sqlock.Lock()
	defer sqlock.Unlock()

	var returnMap = make(map[string]string)
	returnMap["Error"] = ""

	rows, err := s.DB.Query(fmt.Sprintf("select ID, userId, tKey, Password from USERS where userId='%s';", userId))

	if err != nil {
		log.Info().Msgf("select query failed. '%s'", err.Error())
		returnMap["Error"] = err.Error()
		return returnMap
	}
	for rows.Next() {
		var user model.USER
		var tKey string
		var dbId int

		rows.Scan(&dbId, &user.UserId, &tKey, &user.Password)

		returnMap[user.UserId] = user.Password
		returnMap["tKey"] = tKey
	}
	return returnMap
}

func (s *Store) UpdateUser(userId, password string) {
	sqlock.Lock()
	defer sqlock.Unlock()

	_, err := s.DB.Exec(fmt.Sprintf("update USERS set Password='%s' where UserId='%s';", password, userId))
	if err != nil {
		log.Panic().Msgf("update USERS table error - '%s'", err.Error())
	}
}

func (s *Store) DeleteUser(userId string) {
	sqlock.Lock()
	defer sqlock.Unlock()

	_, err := s.DB.Exec(fmt.Sprintf("delete from USERS where UserId='%s';", userId))
	if err != nil {
		log.Panic().Msgf("delete from USERS table error - '%s'", err.Error())
	}
}

func (s *Store) DeleteItems(tKey string) {
	sqlock.Lock()
	defer sqlock.Unlock()

	_, err := s.DB.Exec(fmt.Sprintf("delete from ITEMS where tKey='%s';", tKey))
	if err != nil {
		log.Panic().Msgf("delete from ITEMS table error - '%s'", err.Error())
	}
}

func (s *Store) DeleteItem(dKey string) {
	sqlock.Lock()
	defer sqlock.Unlock()

	_, err := s.DB.Exec(fmt.Sprintf("delete from ITEMS where dKey='%s';", dKey))
	if err != nil {
		log.Panic().Msgf("delete from ITEMS table error - '%s'", err.Error())
	}
}

func (s *Store) SelectItems(tKey string) []map[string]string {
	sqlock.Lock()
	defer sqlock.Unlock()

	var keyMapArray []map[string]string

	rows, err := s.DB.Query(fmt.Sprintf("select tKey, dKey, name, description, value, purchasedate, serialnum from ITEMS where tKey='%s';", tKey))
	if err != nil {
		log.Panic().Msgf("select all keys from ITEMS table error - '%s'", err.Error())
	}
	for rows.Next() {
		var keysMap = make(map[string]string)
		var tKey string
		var dKey string
		var name string
		var description string
		var value string
		var purchasedate string
		var serialnum string

		rows.Scan(&tKey, &dKey, &name, &description, &value, &purchasedate, &serialnum)
		keysMap["dKey"] = dKey
		keysMap["name"] = name
		keysMap["description"] = description
		keysMap["value"] = value
		keysMap["purchasedate"] = purchasedate
		keysMap["serialnum"] = serialnum
		keyMapArray = append(keyMapArray, keysMap)
	}
	return keyMapArray
}

func (s *Store) SelectUserdKeyList(tKey string) []string {
	sqlock.Lock()
	defer sqlock.Unlock()

	var dKeyList []string
	rows, err := s.DB.Query(fmt.Sprintf("select dKey from ITEMS where tKey='%s';", tKey))
	if err != nil {
		log.Panic().Msgf("select tKey from USERS table error - '%s'", err.Error())
	}
	for rows.Next() {
		var dKey string
		rows.Scan(&dKey)
		dKeyList = append(dKeyList, dKey)
	}
	return dKeyList
}

func (s *Store) SelectTDkeyList() []string {
	sqlock.Lock()
	defer sqlock.Unlock()

	var tKeyCodeList []string
	rows, err := s.DB.Query("select tKey from USERS")

	if err != nil {
		log.Panic().Msgf("select tKey from USERS table error - '%s'", err.Error())
	}
	for rows.Next() {
		var tKey string
		rows.Scan(&tKey)
		tKeyCodeList = append(tKeyCodeList, tKey)
	}
	rows, err = s.DB.Query("select dKey from ITEMS")

	if err != nil {
		log.Panic().Msgf("select dKey from ITEMS table error - '%s'", err.Error())
	}
	for rows.Next() {
		var dKey string
		rows.Scan(&dKey)
		tKeyCodeList = append(tKeyCodeList, dKey)
	}
	return tKeyCodeList
}
