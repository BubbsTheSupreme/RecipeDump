package main

import (
	"database/sql"
	"encoding/json"
	"flag"
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

func QueryDb(db *sql.DB, name string) (Recipe, error) {
	rec := Recipe{}

	rows, err := db.Query(`
		SELECT name, ingredients, steps FROM recipes WHERE name = $1 
	`, name)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() { //Boolean returning function that checks if there is another row and points rows.Scan() to it
		err := rows.Scan(&rec.Name, pq.Array(&rec.Ingredients), &rec.Steps)
		if err != nil {
			panic(err)
		}
	}

	return rec, nil
}

func main() {

	var help_flag = flag.Bool("help", false, "Show Help")
	var import_flag string = "None"
	var export_flag string = "None"
	var create_flag string = "None"
	var recipe_name_flag string = "None"

	flag.StringVar(&import_flag, "import", "None", "Imports data from JSON file to PostgreSQL. Arguement value is the name of the file")
	flag.StringVar(&export_flag, "export", "None", "Exports data from PostgreSQL to JSON File. Arguement value is the name of the file. recipe name needs to be provided.")
	flag.StringVar(&create_flag, "create", "None", "Creates the database on the PostgreSQL Server.")
	flag.StringVar(&recipe_name_flag, "recipe", "None", "Gets recipe name as input for export.")

	flag.Parse()

	if *help_flag {
		flag.Usage()
		os.Exit(0)
	}

	data := Recipe{}
	var psqlData string = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlData)
	if err != nil {
		panic(err)
	}

	if create_flag != "None" {
		CreateDb(db)
	}
	if import_flag != "None" {
		var err = ImportJSON(import_flag, &data)
		if err != nil {
			panic(err)
		}
		InsertDb(db, data)
	}
	if export_flag != "None" || recipe_name_flag != "None" {
		var recData Recipe
		recData, err = QueryDb(db, recipe_name_flag)
		if err != nil {
			panic(err)
		}
		var json, err = ExportJSON(recData)
		if err != nil {
			panic(err)
		}

		os.WriteFile(export_flag, []byte(json), 0644)

	}

	defer db.Close()
}
