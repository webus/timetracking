package main

import "fmt"
import "log"
import "time"
import "strconv"
import "database/sql"

// get today workday info about project
func today(project_name string) {
        db,err := sql.Open("postgres",connection_string())
        startWorkDay, endWorkDay := get_workday()
        if err != nil {
                log.Fatal(err)
        } else {
                fmt.Println("")
                query := "SELECT wl.uid,st.state_name,wl.action_time " +
                         "FROM worklog wl " +
                         "INNER JOIN states st on st.uid = wl.state_id " +
                         "WHERE wl.action_time >= $2 AND wl.action_time <= $3 " +
                         "AND wl.project_id = (SELECT uid FROM projects WHERE project_name = $1) " +
                         "ORDER BY wl.action_time ASC;"
                rows, err := db.Query(query, project_name, startWorkDay, endWorkDay)
                if err != nil {
                        log.Fatal(err)
                }
                all_data := []TodayStruct{}
                for rows.Next() {
                        var uid int
                        var state_name string
                        var action_time time.Time
                        if err := rows.Scan(&uid,&state_name,&action_time); err != nil {
                                log.Fatal(err)
                        }
                        all_data = append(all_data, TodayStruct{action_time:action_time, state_name:state_name})
                }
                db.Close()
                var startDate time.Time
                var endDate time.Time
                var nullTime time.Time
                var all_time float64
                for _,v := range all_data {
                        if v.state_name == "start" {
                                startDate = v.action_time
                        }
                        if v.state_name == "stop" {
                                endDate = v.action_time
                                if startDate != nullTime {
                                        delta := endDate.Sub(startDate)
                                        hours := delta.Hours()
                                        hours_s := fmt.Sprintf("%.2f", hours)
                                        hours_f, _ := strconv.ParseFloat(hours_s, 2)
                                        fmt.Printf("%v - %v = %v (h)\n", startDate, endDate, hours_s)
                                        all_time += hours_f
                                }
                        }
                }
                fmt.Println("")
                fmt.Printf("Summary           : %v (hours)\n", round2string(all_time))
                rate, currency_name := get_rate(project_name, time.Now())
                fmt.Printf("Rate              : %s (%s) (per hour) \n",round2string(rate), currency_name)
                fmt.Printf("Money             : %s (%s)\n", round2string((all_time * rate)), currency_name)
        }
}

func full_today() {
        fmt.Println(" === ")
        startDate, endDate := get_workday()
        fmt.Printf("Current workday : %v - %v\n", startDate, endDate)
        projects := project_active_in_time(startDate, endDate)
        var all_time float64
        all_time = 0
        var all_sum float64
        all_sum = 0
        // calc time off all selected projects
        for _, project_name := range projects {
                project_time := round2float(today_func(project_name, startDate, endDate))
                project_rate, currency := get_rate(project_name, startDate)
                project_sum := round2float(project_time * project_rate)
                all_time += project_time
                all_sum += project_sum
                fmt.Println(" --- ")
                fmt.Printf("Project    : %s\n",project_name)
                fmt.Printf("Hours      : %v (h)\n",project_time)
                fmt.Printf("Money      : %v (%s)\n",project_sum, currency)
                fmt.Println(" --- ")
        }
        fmt.Printf("All time  : %v (h)\n", all_time)
        fmt.Println(" === ")
}
