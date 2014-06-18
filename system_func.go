package main

import "os"
import "path"
import "fmt"
import "log"
import "strconv"
import "io/ioutil"
import "syscall"
import "os/exec"
import "encoding/json"
import "database/sql"
import "bufio"

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

func get_connection() (*sql.DB) {
        db,err := sql.Open("postgres",connection_string())
        if err != nil {
                log.Fatal(err)
        } else {
                return db
        }
        return nil
}

// get text from VI editor
func get_from_editor() string {
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
        return string(b)
}
