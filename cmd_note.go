package main

import "log"
import "fmt"

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
        db := get_connection()
        query := "SELECT nt.note_text " +
                 "FROM notes nt " +
                 "INNER JOIN projects prj ON prj.uid = nt.project_id " +
                 "WHERE prj.project_name = $1"
        rows, err := db.Query(query,project_name)
        if err != nil {
                log.Fatal(err)
        }
        for rows.Next() {
                var note string
                err := rows.Scan(&note)
                if err != nil {
                        log.Fatal(err)
                }
                fmt.Println(note)
        }
}
