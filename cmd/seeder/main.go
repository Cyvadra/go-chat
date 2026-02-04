package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gzydong/go-chat/config"
	"github.com/gzydong/go-chat/internal/entity"
	"github.com/gzydong/go-chat/internal/pkg/encrypt"
	"github.com/gzydong/go-chat/internal/pkg/jwtutil"
	"github.com/gzydong/go-chat/internal/provider"
	"github.com/gzydong/go-chat/internal/repository/model"
	"github.com/samber/lo"
)

type UserCredentials struct {
	UserId int    `json:"user_id"`
	Token  string `json:"token"`
}

func main() {
	// 1. Load Config
	conf := config.New("./config.yaml")

	// 2. Init DB
	db := provider.NewMySQLClient(conf)

	fmt.Println("Starting user seeding...")

	password := encrypt.HashPassword("123456")
	baseMobile := 18800000000
	var credentials []UserCredentials

	for i := 0; i < 2000; i++ {
		mobile := fmt.Sprintf("%d", baseMobile+i)
		nickname := fmt.Sprintf("User%d", i)

		var user model.Users
		var count int64

		// Check if exists
		db.Model(&model.Users{}).Where("mobile = ?", mobile).Count(&count)

		if count > 0 {
			db.Model(&model.Users{}).Where("mobile = ?", mobile).First(&user)
			fmt.Printf("User %s already exists (ID: %d), using existing user\n", mobile, user.Id)
		} else {
			user = model.Users{
				Mobile:    lo.ToPtr(mobile),
				Nickname:  nickname,
				Gender:    model.UsersGenderDefault,
				Password:  password,
				IsRobot:   model.No,
				Status:    model.UsersStatusNormal,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			if err := db.Create(&user).Error; err != nil {
				log.Printf("Failed to create user %s: %v\n", mobile, err)
				continue
			}
			fmt.Printf("Created user %s (ID: %d)\n", mobile, user.Id)
		}

		// Generate Token
		token, err := jwtutil.NewTokenWithClaims(
			[]byte(conf.Jwt.Secret), entity.WebClaims{
				UserId: int32(user.Id),
			},
			func(c *jwt.RegisteredClaims) {
				c.Issuer = entity.JwtIssuerWeb
			},
			jwtutil.WithTokenExpiresAt(time.Duration(conf.Jwt.ExpiresTime)*time.Second),
		)

		if err != nil {
			log.Printf("Failed to generate token for user %d: %v\n", user.Id, err)
			continue
		}

		credentials = append(credentials, UserCredentials{
			UserId: user.Id,
			Token:  token,
		})
	}

	// Write to JSON file
	file, err := os.Create("./users.json")
	if err != nil {
		log.Fatalf("Failed to create users.json: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(credentials); err != nil {
		log.Fatalf("Failed to encode credentials: %v", err)
	}

	fmt.Println("Seeding completed. Credentials saved to users.json")
}
