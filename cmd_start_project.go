package main

import "fmt"
import "log"
import "database/sql"

// start and stop work on project
func project_action(action_name string, project_name string, action_comment string) {
        db, err := sql.Open("postgres",connection_string())
        if err != nil {
                log.Fatal(err)
        } else {
                build_sql := fmt.Sprintf("SELECT %s_project($1,$2)", action_name)
                rows, err := db.Query(build_sql, project_name, action_comment)
                if err != nil {
                        log.Fatal(err)
                } else {
                        for rows.Next() {
                                var start_result int
                                if err := rows.Scan(&start_result); err != nil {
                                        log.Fatal(err)
                                }
                                if start_result == 0 {
                                        fmt.Printf("project [%s] action %s. ok.\n", project_name, action_name)
                                } else {
                                        fmt.Printf("!!! project [%s] can't %s, because it %s already !!!\n", project_name, action_name, action_name)
                                }
                        }
                }
        }
}
