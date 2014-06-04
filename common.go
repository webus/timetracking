package main

import "time"

var current_project string = "none"

type DbConfiguration struct {
        UserName string // json: userName
        Password string // json: password
        HostName string // json: hostName
        DatabaseName string // json: databaseName
}

// struct for calc time
type TodayStruct struct {
        action_time time.Time
        state_name string
}
