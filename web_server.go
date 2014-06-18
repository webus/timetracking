package main

//import "io"
import "log"
import "net/http"
import "html/template"

func web_hello(res http.ResponseWriter, req *http.Request) {
        res.Header().Set("Content-Type","text/html")
        tmpl, _ := template.New("name").Parse(`
<html>
<head>
  <title>timetracking</title>
  <link rel="stylesheet" href="//netdna.bootstrapcdn.com/bootstrap/3.1.1/css/bootstrap.min.css">
  <script src="//code.jquery.com/jquery-1.11.0.min.js"></script>
  <script src="//code.jquery.com/jquery-migrate-1.2.1.min.js"></script>
  <script src="//netdna.bootstrapcdn.com/bootstrap/3.1.1/js/bootstrap.min.js"></script>
</head>
<body>
  <div class='container'>

    <div class='row' style='padding-top: 30px;'>
      <div class='col-md-12 text-center'>
        <img src='https://github.com/webus/timetracking/raw/master/docs/img/logo.png' />
      </div>
    </div>

    <div class='row'>
      <div class='col-md-12'>
        <p class='lead'></p>
      </div>
    </div>
  </div>
</body>
</html>
`)
        err := tmpl.ExecuteTemplate(res,"T","TEST")
        if err != nil {
                log.Fatal(err)
        }
}
