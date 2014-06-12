package main

import "log"
import "fmt"
import "database/sql"

func get_connection() (*sql.DB) {
        db,err := sql.Open("postgres",connection_string())
        if err != nil {
                log.Fatal(err)
        } else {
                return db
        }
        return nil
}

// add note to project
func add_note(project_name string, note_text string) {
        db := get_connection()
        query := "INSERT INTO notes(project_id,note_text) VALUES((SELECT uid FROM projects WHERE project_name = $1), $2)"
        _, err := db.Exec(query, project_name, note_text)
        if err != nil {
                log.Fatal(err)
        } else {
                fmt.Printf("Got new note\n")
        }
        db.Close()
}

//
func get_note(project_name string) {
        //

}
