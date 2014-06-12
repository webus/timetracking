package main

import "log"
import "fmt"
import "time"
import "strconv"
import "database/sql"

// helper function to get rate for project and for special time
func get_rate(project_name string, to_time time.Time) (float64, string) {
        db,err := sql.Open("postgres",connection_string())
        if err != nil {
                log.Fatal(err)
        } else {
                query := "select rt.rate, cr.currency_name from rates rt " +
                         "inner join projects prj on prj.uid = rt.project_id " +
                         "inner join currency cr on cr.uid = rt.currency_id " +
                         "where rt.from_date < $1 and to_date > $1 and prj.project_name = $2"
                rows, err := db.Query(query, to_time, project_name)
                if err != nil {
                        log.Fatal(err)
                }
                for rows.Next() {
                        var rate float64
                        var currency_name string
                        if err := rows.Scan(&rate, &currency_name); err != nil {
                                log.Fatal(err)
                        }
                        return rate, currency_name
                }
                db.Close()
        }
        return 0, ""
}

func get_state_of_project(project_name string) (
        state_name string,
        action_time time.Time,
        action_comment string,
        hours_left string,
        rate string,
) {
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
                        delta := time.Since(action_time)
                        hours := delta.Hours()
                        hours_s := fmt.Sprintf("%.2f", hours)
                        rate, _ := get_rate(project_name, time.Now())
                        return state_name, action_time, action_comment, hours_s, round2string(rate)
                }
                db.Close()
        }
        return
}

func get_comment_by_period(project_name string, startDate time.Time, endDate time.Time) {
        db,err := sql.Open("postgres",connection_string())
        if err != nil {
                log.Fatal(err)
        } else {
                query := "SELECT wl.uid,wl.action_time, wl.action_comment " +
                         "FROM worklog wl " +
                         "WHERE wl.action_time >= $2 AND wl.action_time <= $3 " +
                         "AND wl.project_id = (SELECT uid FROM projects WHERE project_name = $1) " +
                         "ORDER BY wl.action_time ASC;"
                rows, err := db.Query(query, project_name, startDate, endDate)
                if err != nil {
                        log.Fatal(err)
                }
                for rows.Next() {
                        var uid int
                        var action_comment string
                        var action_time time.Time
                        if err := rows.Scan(&uid, &action_time, &action_comment); err != nil {
                                log.Fatal(err)
                        }
                        fmt.Printf("### ID : %v\n", uid)
                        fmt.Printf("Date : %v\n", action_time)
                        fmt.Printf("Comment:\n %s\n", action_comment)
                }
                db.Close()
        }
}

// how much time spend to project from start date to end date
func today_func(project_name string, startDate time.Time, endDate time.Time) float64 {
        db,err := sql.Open("postgres",connection_string())
        var all_time float64
        if err != nil {
                log.Fatal(err)
        } else {
                query := "SELECT wl.uid,st.state_name,wl.action_time " +
                         "FROM worklog wl " +
                         "INNER JOIN states st on st.uid = wl.state_id " +
                         "WHERE wl.action_time >= $2 AND wl.action_time <= $3 " +
                         "AND wl.project_id = (SELECT uid FROM projects WHERE project_name = $1) " +
                         "ORDER BY wl.action_time ASC;"
                rows, err := db.Query(query, project_name, startDate, endDate)
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
                                        all_time += hours_f
                                }
                        }
                }
        }
        return all_time
}



// get today workday info about project
func by_day(project_name string, startDate time.Time, endDate time.Time) {
        db,err := sql.Open("postgres",connection_string())
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
                rows, err := db.Query(query, project_name, startDate, endDate)
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

// start or stop workday
func workday_action(action_name string) {
        db, err := sql.Open("postgres",connection_string())
        if err != nil {
                log.Fatal(err)
        } else {
                build_sql := fmt.Sprintf("SELECT %s_workday()", action_name)
                rows, err := db.Query(build_sql)
                if err != nil {
                        log.Fatal(err)
                } else {
                        for rows.Next() {
                                var start_result int
                                if err := rows.Scan(&start_result); err != nil {
                                        log.Fatal(err)
                                }
                                if start_result == 0 {
                                        fmt.Printf("workday is %s. ok.\n", action_name)
                                } else {
                                        fmt.Printf("!!! workday can't be %s !!!\n", action_name)
                                }
                        }
                }
        }
}

//
func project_active_in_time(startDate time.Time, endDate time.Time) []string {
        var projects []string
        db, err := sql.Open("postgres",connection_string())
        if err != nil {
                log.Fatal(err)
        } else {
                build_sql := "SELECT DISTINCT prj.project_name FROM worklog wl " +
                             "INNER JOIN projects prj ON prj.uid = wl.project_id " +
                             "WHERE wl.action_time >= $1 AND wl.action_time <= $2"
                rows, err := db.Query(build_sql, startDate, endDate)
                if err != nil {
                        log.Fatal(err)
                } else {
                        for rows.Next() {
                                var project_name string
                                if err := rows.Scan(&project_name); err != nil {
                                        log.Fatal(err)
                                }
                                projects = append(projects, project_name)
                        }
                }
        }
        return projects
}

// search start date and end date of work day
func search_workday(state_name string) (state_name_ret string, action_time time.Time) {
        db,err_db := sql.Open("postgres",connection_string())
        if state_name == "" {
                query := "SELECT st.state_name, wd.action_time FROM workday wd " +
                         "INNER JOIN states st on st.uid = wd.state_id " +
                         "ORDER BY wd.action_time DESC LIMIT 1"
                if err_db != nil {
                        log.Fatal(err_db)
                } else {
                        rows, err := db.Query(query)
                        if err != nil {
                                log.Fatal(err)
                        }
                        for rows.Next() {
                                var state_name_ret string
                                var action_time time.Time
                                if err := rows.Scan(&state_name_ret, &action_time); err != nil {
                                        log.Fatal(err)
                                }
                                return state_name_ret, action_time
                        }
                        db.Close()
                }
        } else {
                query := "SELECT st.state_name, wd.action_time FROM workday wd " +
                         "INNER JOIN states st on st.uid = wd.state_id " +
                         "WHERE st.state_name = $1 " +
                         "ORDER BY wd.action_time DESC LIMIT 1"
                if err_db != nil {
                        log.Fatal(err_db)
                } else {
                        rows, err := db.Query(query, state_name)
                        if err != nil {
                                log.Fatal(err)
                        }
                        for rows.Next() {
                                var state_name_ret string
                                var action_time time.Time
                                if err := rows.Scan(&state_name_ret, &action_time); err != nil {
                                        log.Fatal(err)
                                }
                                return state_name_ret, action_time
                        }
                        db.Close()
                }
        }
        return "", time.Now()
}

// print start end end date of current workday
func get_workday() (time.Time, time.Time) {
        var beginDate time.Time
        var endDate time.Time
        state_name, action_time := search_workday("")
        if state_name == "stop" {
                endDate = action_time
                state_name, action_time = search_workday("start")
                beginDate = action_time
        } else if state_name == "start" {
                beginDate = action_time
                maxDate, _ := time.Parse("2006-Jan-02", "2222-Jan-01")
                endDate = maxDate
        }
        return beginDate, endDate
}



func full_by_date(projectName string, startDate time.Time, endDate time.Time) {
        project_time := round2float(today_func(projectName, startDate, endDate))
        project_rate, currency := get_rate(projectName, startDate)
        project_sum := round2float(project_time * project_rate)
        // calc time off all selected projects
        get_comment_by_period(projectName, startDate, endDate)
        fmt.Println(" === ")
        fmt.Printf("Project : %s\n", projectName)
        fmt.Printf("Date    : %v - %v\n", startDate, endDate)
        fmt.Printf("Hours   : %v (h)\n", project_time)
        fmt.Printf("Rate    : %v (%s) (hour)\n", project_rate, currency)
        fmt.Printf("Sum     : %v\n", project_sum)
        fmt.Println(" === ")
}
