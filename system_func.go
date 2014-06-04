package main

import "os"
import "path"
import "fmt"
import "strconv"
import "encoding/json"
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
