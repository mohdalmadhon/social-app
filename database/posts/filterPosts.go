package posts

import (
	"database/sql"
	"log"
	"social/database/profiles"
	"social/database/users"
	"social/internal/models"
	"strconv"
	"strings"
)

func FilterPosts(db *sql.DB, posts *[]models.Post, userID, targetID int) ([]models.Post, error) {
	var returnPosts []models.Post

	for _, p := range *posts {
		if p.GroupId == nil {
			if p.Public == 1 {
				returnPosts = append(returnPosts, p)
				continue
			}

			if p.Private == 1 {
				isFollower, err := users.IsFollowing(db, userID, targetID)
				if err != nil {
					continue
				}

				if isFollower {
					returnPosts = append(returnPosts, p)
				}
			}

			continue
		}

		groupID := *p.GroupId

		if groupID == 0 {
			returnPosts = append(returnPosts, p)
			continue
		}

		if groupID == -1 {
			isFollower, err := profiles.CheckFollower(db, userID, targetID)
			if err != nil {
				return nil, err
			}

			if isFollower == 1 {
				returnPosts = append(returnPosts, p)
			}

			continue
		}

		log.Println("group post")
		var users string

		err := db.QueryRow(
			`SELECT users FROM user_posts_groups WHERE id = ?`,
			groupID,
		).Scan(&users)
		
		log.Printf("users %s", users)
		if err != nil {
			log.Println("hereerr")
			if err == sql.ErrNoRows {
				continue
			}

			return nil, err
		}

		isMember := false

		log.Println(strings.Split(users, ":"))
		log.Println(userID)
		for _, idStr := range strings.Split(users, ":") {
			if idStr == "" {
				continue
			}

			id, err := strconv.Atoi(idStr)
			if err != nil {
				return nil, err
			}

			if id == userID {
				isMember = true
				break
			}
		}

		if isMember {
			returnPosts = append(returnPosts, p)
		}
	}

	return returnPosts, nil
}
