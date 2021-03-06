package main

import (
        _ "github.com/lib/pq"
        "database/sql"
        "fmt"
        "bufio"
        "os"
        "log"
        "time"
        "strconv"
        "io/ioutil"
        "syscall"
        "os/exec"
        "encoding/json"
        "path"
)

var current_project string = "none"

/*
  Example: ~/.timetracking.json
  {
        "userName": "postgres",
        "password": "secret",
        "hostName": "localhost",
        "databaseName": "timetracking"
   }
*/
type DbConfiguration struct {
        UserName string // json: userName
        Password string // json: password
        HostName string // json: hostName
        DatabaseName string // json: databaseName
}

func connection_string() string {
        filename := path.Join(os.Getenv("HOME"),".timetracking.json")
        if _, err := os.Stat(filename); os.IsNotExist(err) {
                fmt.Printf("no such file or directory: %s\n", filename)
                os.Exit(1)
        } else {
                file, err := os.Open(filename)
                if err != nil {
                        fmt.Println(err)
                }
                var conf DbConfiguration
                jsonParser := json.NewDecoder(file)
                if err = jsonParser.Decode(&conf); err != nil {
                        fmt.Println(err.Error())
                }
                conn := fmt.Sprintf("postgres://%s:%s@%s/%s", conf.UserName, conf.Password, conf.HostName, conf.DatabaseName)
                return conn
        }
        return ""
}

// helper function to ready string from stdin
func read_string() string {
        bio := bufio.NewReader(os.Stdin)
        line, _, _ := bio.ReadLine()
        return string(line)
}

// round float64 to 0.1
func round2string(num float64) string {
        return fmt.Sprintf("%.2f", num)
}

// round float64
func round2float(num float64) float64 {
        num_s := fmt.Sprintf("%.2f", num)
        num_f, _ := strconv.ParseFloat(num_s, 2)
        return num_f
}

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

func get_active_porjects() {
        projects_list := available_projects_list()
        for index, element := range projects_list {
                state_name, _, _, _, _ := get_state_of_project(element)
                if state_name == "start" {
                        fmt.Printf("%v. - %s\n",index + 1,element)
                }
        }
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

// struct for calc time
type TodayStruct struct {
        action_time time.Time
        state_name string
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

// add note to project
func add_note(project_name string, note_text string) {
        db,err := sql.Open("postgres",connection_string())
        if err != nil {
                log.Fatal(err)
        } else {
                query := "INSERT INTO notes(project_id,note_text) VALUES((SELECT uid FROM projects WHERE project_name = $1), $2)"
                _, err := db.Exec(query, project_name, note_text)
                if err != nil {
                        log.Fatal(err)
                } else {
                        fmt.Printf("Got new note\n")
                }
                db.Close()
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


func main() {
        time_now := time.Now()
        time_now_str := time_now.Format("2006-01-02 15:04:01")
        command_len := len(os.Args)
        if command_len > 1 {
                command := os.Args[1]
                // start project
                if command == "start" {
                        if command_len == 4 {
                                project_name := os.Args[2]
                                action_comment := os.Args[3]
                                if action_comment == "-v" {
                                        f, err := ioutil.TempFile("", "timetrack")
                                        if err != nil { panic(err) }
                                        defer syscall.Unlink(f.Name())
                                        cmd := exec.Command("vi", f.Name())
                                        cmd.Stdin = os.Stdin
                                        cmd.Stdout = os.Stdout
                                        cmd.Stderr = os.Stderr
                                        err_file := cmd.Start()
                                        if err_file != nil {
                                                log.Fatal(err)
                                        }
                                        err_file = cmd.Wait()
                                        if err_file != nil {
                                                log.Fatal(err)
                                        }
                                        b, err_file2 := ioutil.ReadFile(f.Name())
                                        if err_file2 != nil {
                                                log.Fatal(err)
                                        }
                                        action_comment = string(b)
                                }
                                fmt.Printf("%s : trying start %s\n", time_now_str, project_name)
                                fmt.Println("with comment")
                                project_action("start", project_name, action_comment)
                        }
                        if command_len == 3 {
                                project_name := os.Args[2]
                                fmt.Printf("%s : trying start %s\n", time_now_str, project_name)
                                project_action("start", project_name, "")
                        }
                        if command_len == 2 {
                                fmt.Println("type: timetrack start [project_name] ['comment to this action']")
                                fmt.Println("comment is optional")
                        }
                }
                // stop project
                if command == "stop" {
                        if command_len == 3{
                                project_name := os.Args[2]
                                fmt.Printf("%s : trying stop %s\n", time_now_str, project_name)
                                project_action("stop", project_name, "")
                        }
                        if command_len == 4 {
                                project_name := os.Args[2]
                                action_comment := os.Args[3]
                                if action_comment == "-v" {
                                        f, err := ioutil.TempFile("", "timetrack")
                                        if err != nil { panic(err) }
                                        defer syscall.Unlink(f.Name())
                                        cmd := exec.Command("vi", f.Name())
                                        cmd.Stdin = os.Stdin
                                        cmd.Stdout = os.Stdout
                                        cmd.Stderr = os.Stderr
                                        err_file := cmd.Start()
                                        if err_file != nil {
                                                log.Fatal(err)
                                        }
                                        err_file = cmd.Wait()
                                        if err_file != nil {
                                                log.Fatal(err)
                                        }
                                        b, err_file2 := ioutil.ReadFile(f.Name())
                                        if err_file2 != nil {
                                                log.Fatal(err)
                                        }
                                        action_comment = string(b)
                                }
                                fmt.Printf("%s : trying stop %s\n", time_now_str, project_name)
                                fmt.Println("with comment")
                                project_action("stop", project_name, action_comment)
                        }
                        if command_len == 2 {
                                fmt.Println("type: timetrack stop [project_name] ['comment to this action']")
                                fmt.Println("comment is optional")
                        }
                }
                if command == "list" {
                        available_projects()
                }
                if command == "state" {
                        if command_len == 2 {
                                fmt.Printf("get current active projects\n")
                                get_active_porjects()
                        }
                        if command_len == 3 {
                                project_name := os.Args[2]
                                get_state(project_name)
                        }
                }
                if command == "today" {
                        if command_len == 3 {
                                project_name := os.Args[2]
                                fmt.Printf("%s : trying print today info about %s\n", time_now_str, project_name)
                                today(project_name)
                        }
                        if command_len == 2 {
                                full_today()
                        }
                }
                if command == "note" {
                        if command_len == 4 {
                                project_name := os.Args[2]
                                note_text := os.Args[3]
                                add_note(project_name, note_text)

                        }
                }
                if command == "workday" {
                        if command_len == 3 {
                                action_name := os.Args[2]
                                workday_action(action_name)
                        }
                        if command_len == 2 {
                                startWorkDay, endWorkDay := get_workday()
                                fmt.Printf("Workday: %v - %v\n", startWorkDay, endWorkDay)
                        }
                }
                if command == "summary" {
                        if command_len == 5 {
                                project_name := os.Args[2]
                                date_from := os.Args[3]
                                date_to := os.Args[4]
                                loc, _ := time.LoadLocation("Europe/Moscow")
                                date_from_p, _ := time.ParseInLocation("02.01.2006", date_from, loc)
                                date_to_p, _ := time.ParseInLocation("02.01.2006", date_to, loc)
                                fmt.Printf("%s %v %v \n", project_name, date_from_p, date_to_p)
                                full_by_date(project_name, date_from_p, date_to_p)
                        } else {
                                fmt.Println("ERROR: not all parameters use")
                                fmt.Println("For example: timetrack summary [project_name] [from_date] [to_date]")
                                fmt.Println("Example: timetrack summary myproject 01.03.2013 05.03.2013")
                        }
                }
                if command == "by-day" {
                        if command_len == 5 {
                                project_name := os.Args[2]
                                date_from := os.Args[3]
                                date_to := os.Args[4]
                                loc, _ := time.LoadLocation("Europe/Moscow")
                                date_from_p, _ := time.ParseInLocation("02.01.2006", date_from, loc)
                                date_to_p, _ := time.ParseInLocation("02.01.2006", date_to, loc)
                                fmt.Printf("%s %v %v \n", project_name, date_from_p, date_to_p)
                                by_day(project_name, date_from_p, date_to_p)
                        } else {
                                fmt.Println("ERROR: not all parameters use")
                                fmt.Println("For example: timetrack by-day [project_name] [from_date] [to_date]")
                                fmt.Println("Example: timetrack ny-daymyproject 01.03.2013 05.03.2013")
                        }
                }
        } else {
                fmt.Println("timetracking")
                fmt.Println()
                fmt.Println("available commands:")
                fmt.Println("start   [project_name]                       - start time track for project")
                fmt.Println("stop    [project_name]                       - stop time track for project")
                fmt.Println("list                                         - list all projects")
                fmt.Println("state   [project_name]                       - current state of project")
                fmt.Println("today   [project_name]                       - timelog of project for current workday")
                fmt.Println("note    [project_name]                       - add some note to project")
                fmt.Println("workday [start / stop]                       - start or stop workday")
                fmt.Println("summary [project_name] [date_from] [date_to] - list all timelog between dates with all comments")
                fmt.Println("by-day  [project_name] [date_from] [date_to] - list all timelog between dates and summary")
        }
}
