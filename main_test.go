package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	log.Println("Do stuff BEFORE the tests!")
	exitVal := m.Run()
	log.Println("Do stuff AFTER the tests!")

	os.Exit(exitVal)
}

func TestSuccessFetchUserByIDOnylFromCache(t *testing.T) {
	server, _ := miniredis.Run()
	defer server.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	userHash := map[string]interface{}{
		"id":    "1",
		"name":  "john done cache",
		"email": "john_doe@example.com",
	}
	userJSON, err := json.Marshal(userHash)
	if err != nil {
		fmt.Printf("Failed to marshal user JSON: %v", err)
	}
	if err := server.Set("1", string(userJSON)); err != nil {
		fmt.Printf("Failed to set user data in mock Redis server: %v", err)
	}

	ctx := context.Background()

	wg.Add(1)
	go fetchUserByID(ctx, rdb, "1", respch, wg, db)
	wg.Wait()

	expectedJSON := `{"id": "1", "name": "john done cache", "email": "john_doe@example.com"}`
	assert.JSONEq(t, expectedJSON, <-respch)
}

func TestSuccessFetchUserByIDOnylFromDB(t *testing.T) {
	server, _ := miniredis.Run()
	defer server.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectExec("INSERT INTO users (id, name, email) VALUES (?, ?, ?)").WithArgs(1, "John Doe", "john@example.com").WillReturnResult(sqlmock.NewResult(1, 1))

	_, err = db.Exec("INSERT INTO users (id, name, email) VALUES (?, ?, ?)", 1, "John Doe", "john@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	rows := sqlmock.NewRows([]string{"id", "name", "email"}).AddRow(1, "John Doe", "john@example.com")
	mock.ExpectQuery("SELECT * FROM users WHERE id =?").WithArgs(1).WillReturnRows(rows)

	ctx := context.Background()

	wg.Add(1)
	go fetchUserByID(ctx, rdb, "1", respch, wg, db)
	wg.Wait()

	assert.Equal(t, 1, len(respch))
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}
