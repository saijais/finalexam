package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type customer struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

var db *sql.DB

func init() {
	var err error
	link := os.Getenv("DATABASE_URL")
	db, err = sql.Open("postgres", link)
	// db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	//db, err = sql.Open("postgres", "postgres://doabpkaa:tN4h2VoTnhh2mYPos_NRayICvXKlX41f@satao.db.elephantsql.com:5432/doabpkaa")
	if err != nil {
		log.Fatal(err)
	}

	//	createTb := `CREATE TABLE IF NOT EXISTS customers (
	//		   id SERIAL PRIMARY KEY,
	//         name  TEXT,
	//         email TEXT,
	//		   status TEXT
	//	);`

	//	_, err = db.Exec(createTb)
	//	if err != nil {
	//		log.Fatal("can't create table customers ", err)
	//	}

	//	fmt.Println("create table customers success")

}

func createCustomersHandler(c *gin.Context) {
	t := customer{}
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	row := db.QueryRow("INSERT INTO customers (name, email, status) values ($1, $2, $3)  RETURNING id", t.Name, t.Email, t.Status)

	err := row.Scan(&t.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, t)
}

func getCustomersHandler(c *gin.Context) {
	status := c.Query("status")

	stmt, err := db.Prepare("SELECT id, name, email, status FROM customers")
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	rows, err := stmt.Query()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	customers := []customer{}
	for rows.Next() {
		t := customer{}

		err := rows.Scan(&t.ID, &t.Name, &t.Email, &t.Status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}

		customers = append(customers, t)
	}

	tt := []customer{}

	for _, item := range customers {
		if status != "" {
			if item.Status == status {
				tt = append(tt, item)
			}
		} else {
			tt = append(tt, item)
		}
	}

	c.JSON(http.StatusOK, tt)
}

func getCustomersByIDHandler(c *gin.Context) {
	id := c.Param("id")

	stmt, err := db.Prepare("SELECT id, name, email, status FROM customers where id=$1")
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	row := stmt.QueryRow(id)

	t := &customer{}

	err = row.Scan(&t.ID, &t.Name, &t.Email, &t.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, t)
}

func updateCustomersHandler(c *gin.Context) {
	id := c.Param("id")
	stmt, err := db.Prepare("SELECT id, name, email, status FROM customers where id=$1")
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	row := stmt.QueryRow(id)

	t := &customer{}

	err = row.Scan(&t.ID, &t.Name, &t.Email, &t.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	if err := c.ShouldBindJSON(t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stmt, err = db.Prepare("UPDATE customers SET name=$2, email=$3, status=$4 WHERE id=$1;")
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	if _, err := stmt.Exec(id, t.Name, t.Email, t.Status); err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, t)
}

func deleteCustomersHandler(c *gin.Context) {
	id := c.Param("id")
	stmt, err := db.Prepare("DELETE FROM customers WHERE id = $1")
	if err != nil {
		log.Fatal("can't prepare delete statement", err)
	}

	if _, err := stmt.Exec(id); err != nil {
		log.Fatal("can't execute delete statment", err)
	}

	c.JSON(http.StatusOK, "customer deleted.")
}

func authMiddleware(c *gin.Context) {
	fmt.Println("start #middleware")
	token := c.GetHeader("Authorization")
	//  if token != "Bearer token2019" {
	if token != "token2019" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "you don't have the right!!"})
		c.Abort()
		return
	}

	c.Next()

	fmt.Println("end #middleware")

}
func setupRouter() *gin.Engine {
	r := gin.Default()

	apiV1 := r.Group("/api/v1")

	apiV1.Use(authMiddleware)

	apiV1.GET("/customers", getCustomersHandler)
	apiV1.GET("/customers/:id", getCustomersByIDHandler)
	apiV1.POST("/customers", createCustomersHandler)
	apiV1.PUT("/customers/:id", updateCustomersHandler)
	apiV1.DELETE("/customers/:id", deleteCustomersHandler)

	return r
}

func main() {
	r := setupRouter()
	r.Run(":2019")
}
