package main

import "fmt"
import "log"
import "database/sql"

// get list of available projects
func available_projects_list() []string {
        db,err := sql.Open("postgres",connection_string())
        var projects_list []string
        if err != nil {
                log.Fatal(err)
        } else {
                rows, err := db.Query("select project_name from projects")
                if err != nil {
                        log.Fatal(err)
                }
                for rows.Next() {
                        var name string
                        if err := rows.Scan(&name); err != nil {
                                log.Fatal(err)
                        }
                        projects_list = append(projects_list, name)
                }
                db.Close()
        }
        return projects_list
}

// print list of available projects
func available_projects() {
        projects_list := available_projects_list()
        for index, element := range projects_list {
                fmt.Printf("%v. - %s\n",index + 1,element)
        }
}
