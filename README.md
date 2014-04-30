timetracking
============

Tool for tracking your working time.

Every programmer spends a lot of time in the console. No matter what kind of console, it can be Bash or Zsh. As most programmers are working on an hourly basis. How can consider your time while you work ? 
I faced this problem and I solved it, this solution is here. This is simple utility allows you to consider your working hours effortlessly. You can use it for invoicing clients, or simply for time for yourself.

![timetracking L](docs/img/logo.png "timetracking")

It's easy! Look at this!

```
available commands:
start   [project_name]                       - start time track for project
stop    [project_name]                       - stop time track for project
list                                         - list all projects
state   [project_name]                       - current state of project
today   [project_name]                       - timelog of project for current workday
note    [project_name]                       - add some note to project
workday [start / stop]                       - start or stop workday
summary [project_name] [date_from] [date_to] - list all timelog between dates with all comments
by-day  [project_name] [date_from] [date_to] - list all timelog between dates and summary
```

We can start out workday like this:
```bash
$ timetracking workday start
$ timetracking start my_project "start developing some stuff by task #232"
$ timetracking stop my_project
$ timetracking today
$ timetracking workday stop
```

## How to install ?
First you need to install PostgreSQL on your system. Next create database fro example timetracking. Init this database with init.sql from sql/init.sql. Next create configuration file ~/.timetracking.json

Example of ~/.timetracking.json
```json
{
	"userName": "postgres",
	"password": "secret",
	"hostName": "localhost",
	"databaseName": "timetracking"
}

```

Next install timetracking:
```bash
go get github.com/webus/timetracking
```

Enjoy!


