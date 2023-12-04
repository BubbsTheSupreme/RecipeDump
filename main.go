package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/lib/pq"
)

const (
	host     = ""
	port     = 5432
	user     = ""
	dbname   = "recipedump"
	password = ""
)

type Recipe struct {
	Name        string
	Ingredients []string
	Steps       string
}

func ImportJSON(file string, rec *Recipe) error {
	data, err := os.ReadFile(file)
	if err != nil {
		fmt.Println("cant read file")
		return err
	}

	errr := json.Unmarshal(data, rec)
	if errr != nil {
		fmt.Println("cant unmarshal json")
		return err
	}
	return nil
}

func ExportJSON(rec Recipe) (string, error) {
	jsonData, err := json.Marshal(&rec)
	if err != nil {
		return string(jsonData), err
	}
	return string(jsonData), nil
}

func InsertDb(db *sql.DB, data Recipe) { //accept JSON struct instead

	_, err := db.Exec(`
	INSERT INTO recipes (name, steps, ingredients)
	VALUES (
		$1, $2, $3 
	);`, data.Name, data.Steps, pq.Array(data.Ingredients))

	if err != nil {
		fmt.Println(err)
	}
}

func CreateDb(db *sql.DB) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS recipes(
			"id" SERIAL PRIMARY KEY,
			"name" text,
			"ingredients" text[],
			"steps" text
	);`)

	if err != nil {
		panic(err)
	} else {
		fmt.Println("Created database")
	}
}

func main() {

	var psqlData string = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlData)
	if err != nil {
		panic(err)
	}

	CreateDb(db)

	data := Recipe{}

	ImportJSON("test.json", &data)

	fmt.Println(data)

	InsertDb(db, data)

	defer db.Close()
}
