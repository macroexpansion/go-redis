package db

import (
	"fmt"
	"github.com/go-redis/redis/v9"
)

type User struct {
	Username string `json:"username" binding:"required"`
	Points   int    `json:"points,string"`
	Rank     int    `json:"rank,string"`
}

func (db *Database) SaveUser(user *User) error {
	member := &redis.Z{
		Score:  float64(user.Points),
		Member: user.Username,
	}
	pipe := db.Client.TxPipeline() // redis transaction and pipeline
	pipe.ZAdd(Ctx, "leaderboard", *member)
	rank := pipe.ZRank(Ctx, "leaderboard", user.Username)
	_, err := pipe.Exec(Ctx)
	if err != nil {
		return err
	}
	fmt.Println(rank.Val(), err)
	user.Rank = int(rank.Val())
	return nil
}

func (db *Database) GetUser(username string) (*User, error) {
	pipe := db.Client.TxPipeline()
	score := pipe.ZScore(Ctx, "leaderboard", username)
	rank := pipe.ZRank(Ctx, "leaderboard", username)
	_, err := pipe.Exec(Ctx)
	if err != nil {
		return nil, err
	}
	if score == nil {
		return nil, ErrNil
	}

	return &User{
		Username: username,
		Points:   int(score.Val()),
		Rank:     int(rank.Val()),
	}, nil
}
