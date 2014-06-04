package main

import "fmt"
import "log"
import "time"
import "strconv"
import "database/sql"

func get_active_porjects() {
        projects_list := available_projects_list()
        for index, element := range projects_list {
                state_name, _, _, _, _ := get_state_of_project(element)
                if state_name == "start" {
                        fmt.Printf("%v. - %s\n",index + 1,element)
                }
        }
}

// print current state of project
func get_state(project_name string) {
        db,err := sql.Open("postgres",connection_string())
        if err != nil {
                log.Fatal(err)
        } else {
                query := "select wl.action_time, st.state_name, wl.action_comment from worklog wl " +
                         "inner join projects prj on prj.uid = wl.project_id " +
                         "inner join states st on st.uid = wl.state_id " +
                         "where prj.project_name = $1 " +
                         "order by wl.action_time desc limit 1"
                rows, err := db.Query(query, project_name)
                if err != nil {
                        log.Fatal(err)
                }
                for rows.Next() {
                        var action_time time.Time
                        var state_name string
                        var action_comment string
                        if err := rows.Scan(&action_time, &state_name, &action_comment); err != nil {
                                log.Fatal(err)
                        }
                        fmt.Println(" === ")
                        fmt.Printf("Project           : %s\n", project_name)
                        fmt.Printf("Last action       : %s \n", state_name)
                        fmt.Printf("Action started at : %v \n",action_time)
                        fmt.Printf("Now               : %v \n",time.Now())
                        delta := time.Since(action_time)
                        hours := delta.Hours()
                        hours_s := fmt.Sprintf("%.2f", hours)
                        hours_f, _ := strconv.ParseFloat(hours_s, 2)
                        fmt.Printf("Hours left        : %s (h)\n", hours_s)
                        rate, currency_name := get_rate(project_name, time.Now())
                        fmt.Printf("Rate              : %s (%s) (per hour) \n", round2string(rate), currency_name)
                        fmt.Printf("Money             : %s (%s)\n", round2string((hours_f * rate)), currency_name)
                        fmt.Printf("Comment           : %s\n", action_comment)
                        fmt.Println(" === ")
                }
                db.Close()
        }
}
