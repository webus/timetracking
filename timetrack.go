package main

import (
        _ "github.com/lib/pq"
        "fmt"
        "os"
        "log"
        "time"
        "io/ioutil"
        "syscall"
        "os/exec"
        "net/http"
)

/*
  Example: ~/.timetracking.json
  {
        "userName": "postgres",
        "password": "secret",
        "hostName": "localhost",
        "databaseName": "timetracking"
   }
*/

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
                                note_cmd := os.Args[3]
                                if note_cmd == "new" {
                                        note_text := get_from_editor()
                                        add_note(project_name, note_text)
                                }
                                if note_cmd == "view" {
                                        get_note(project_name)
                                }

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
                if command == "web" {
                        fmt.Println("http://0.0.0.0:9000")
                        http.HandleFunc("/", web_hello)
                        http.ListenAndServe(":9000",nil)
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
