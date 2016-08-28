package main

import (
    "bytes"
    "fmt"
    "log"
    "net/http"
    "os"
    "strconv"
    "time"
    "database/sql"

    "github.com/gin-gonic/gin"
    "github.com/russross/blackfriday"
    _ "github.com/lib/pq"
    "github.com/jinzhu/gorm"    
    _ "github.com/jinzhu/gorm/dialects/postgres"
)

var (
    repeat int
    db     *sql.DB
    db2    *gorm.DB
)

func repeatFunc(c *gin.Context) {
    var buffer bytes.Buffer
    for i := 0; i < repeat; i++ {
        buffer.WriteString("Hello from Go!")
    }
    c.String(http.StatusOK, buffer.String())
}

func dbFunc(c *gin.Context) {
    if _, err := db.Exec("CREATE TABLE IF NOT EXISTS ticks (tick timestamp)"); err != nil {
        c.String(http.StatusInternalServerError,
            fmt.Sprintf("Error creating database table: %q", err))
        return
    }

    if _, err := db.Exec("INSERT INTO ticks VALUES (now())"); err != nil {
        c.String(http.StatusInternalServerError,
            fmt.Sprintf("Error incrementing tick: %q", err))
        return
    }

    rows, err := db.Query("SELECT tick FROM ticks")
    if err != nil {
        c.String(http.StatusInternalServerError,
            fmt.Sprintf("Error reading ticks: %q", err))
        return
    }

    defer rows.Close()
    for rows.Next() {
        var tick time.Time
        if err := rows.Scan(&tick); err != nil {
          c.String(http.StatusInternalServerError,
            fmt.Sprintf("Error scanning ticks: %q", err))
            return
        }
        c.String(http.StatusOK, fmt.Sprintf("Read from DB: %s\n", tick.String()))
    }
}

type User struct {
  gorm.Model
  Name string
}

func dbGormFunc(c *gin.Context) {

 	if !db2.HasTable("users") {
		db2.CreateTable(&User{})
  	}
	
	user := User{Name: "Gilad"}
	db2.Create(&user)

	uu := db2.Last(&user)
   c.String(http.StatusOK, fmt.Sprintf("Read from DB: %s\n", uu))
}

func main() {
    port := os.Getenv("PORT")

    if port == "" {
        log.Fatal("$PORT must be set")
    }

    var err error
    var err2 error

    tStr := os.Getenv("REPEAT")
    repeat, err = strconv.Atoi(tStr)
    if err != nil {
        log.Print("Error converting $REPEAT to an int: %q - Using default", err)
        repeat = 5
    }

    db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatalf("Error opening database: %q", err)
    }

	db2, err2 = gorm.Open("postgres", os.Getenv("DATABASE_URL"))
    if err2 != nil {
        log.Fatalf("Error opening database: %q", err2)
    }

    router := gin.New()
    router.Use(gin.Logger())
    router.LoadHTMLGlob("templates/*.tmpl.html")
    router.Static("/static", "static")

    router.GET("/", func(c *gin.Context) {
        c.HTML(http.StatusOK, "index.tmpl.html", nil)
    })

    router.GET("/mark", func(c *gin.Context) {
        c.String(http.StatusOK, string(blackfriday.MarkdownBasic([]byte("**hi!**"))))
    })

    router.GET("/repeat", repeatFunc)
    router.GET("/db", dbFunc)

    router.Run(":" + port)
}
