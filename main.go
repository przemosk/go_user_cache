package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/go-sql-driver/mysql"
	"github.com/phuslu/log"
)

// User represents user information
type User struct {
	id    string "json:id"
	name  string "json:name"
	email string "json:email"
}

var (
	wg     = &sync.WaitGroup{}
	respch = make(chan string, 1)
)

func fillRedisCacheDatabase(ctx context.Context, redis_c redis.Cmdable, up_to int) error {
fillup:
	for i := 1; i < up_to; i++ {
		user := User{
			id:    fmt.Sprintf("%d", i),
			name:  "John Doe_" + fmt.Sprintf("%d", i),
			email: "john_doe@example.com",
		}

		userHash := map[string]interface{}{
			"id":    user.id,
			"name":  user.name,
			"email": user.email}

		u, err_m := json.Marshal(userHash)

		if err_m != nil {
			fmt.Println(err_m)
		}
		// Set user in Redis
		err := redis_c.Set(ctx, user.id, u, 0).Err()
		if err != nil {
			fmt.Println("Error setting user:", err)
			break fillup
		}
	}
	return nil
}

func fetchDBUserByID(id int64, respch chan string, wg *sync.WaitGroup, db *sql.DB) {
	var user User
	defer wg.Done()

	log.Info().Msg("LOOKING INTO DATABASE")
	row := db.QueryRow("SELECT * FROM users WHERE id =?", id)
	if err := row.Scan(&user.id, &user.name, &user.email); err != nil {
		if err == sql.ErrNoRows {
			log.Error().Msgf("usersById %d: no such user", id)
		}
		log.Error().Msgf("usersById %d: %v", id, err)
	}

	userString := fmt.Sprintf(user.id + " " + user.name + " " + user.email)
	respch <- userString
}

func fetchUserByID(ctx context.Context, redis_c redis.Cmdable, key string, respch chan string, wg *sync.WaitGroup, db *sql.DB) {
	defer wg.Done()
	log.Info().Msg("LOOKING INTO CACHE")

	stringUserID := key
	intUserID, _ := strconv.ParseInt(stringUserID, 10, 64)

	cachedUser, err := redis_c.Get(ctx, stringUserID).Result()
	if err != nil {
		log.Info().Msg("NOT FOUND IN CACHE")
	}

	if cachedUser != "" {
		respch <- cachedUser
	}

	if cachedUser == "" {
		wg.Add(1)
		go fetchDBUserByID(intUserID, respch, wg, db)
	}
}

func main() {
	ctx := context.Background()

	// Initialize Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// Capture connection properties.
	cfg := mysql.Config{
		User:   "root",
		Passwd: "",
		Net:    "tcp",
		Addr:   "127.0.0.1:3306",
		DBName: "user_cache_dev",
	}
	// Get a database handle.
	var err error
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Error().Err(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Error().Err(err)
	}

	// fillRedisCacheDatabase(ctx, rdb, 2)

	wg.Add(1)
	go fetchUserByID(ctx, rdb, "1", respch, wg, db)
	wg.Wait()
	close(respch)

	fmt.Println("USER:", <-respch)

	// for i := 1; i < 10; i++ {
	// 	str := strconv.Itoa(i)

	// 	wg.Add(1)
	// 	go fetchUserByID(ctx, rdb, str, respch, wg, db)
	// 	wg.Wait()

	// 	fmt.Println("USER:", <-respch)

	// }
	// close(respch)

}
