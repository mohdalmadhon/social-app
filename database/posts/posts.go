package posts

import (
	"database/sql"
	"fmt"
	"log"
	"social/internal/models"
	"strconv"
	"strings"
)

func AddPost(db *sql.DB, post models.RegsiterPost) (int, error) {
	tags := func() string {
		values := make([]string, len(post.PeopleTagged))
		for i, v := range post.PeopleTagged {
			values[i] = strconv.Itoa(v)
		}
		return strings.Join(values, ":")
	}

	var result sql.Result
	var err error

	log.Println("groupID", post.GroupID)
	if post.GroupID > 0 {
		var exists int

		err = db.QueryRow(`
            SELECT 1
            FROM user_posts_groups
            WHERE id = ?
        `, post.GroupID).Scan(&exists)

		if err != nil {
			if err == sql.ErrNoRows {
				return 0, fmt.Errorf("group %d does not exist", post.GroupID)
			}
			return 0, err
		}

		result, err = db.Exec(`
            INSERT INTO posts (
                user_id,
                content,
                image_path,
                allow_comments,
                location,
                group_id,
                public,
                private,
                tags
            )
            VALUES (?, ?, ?, ?, ?, ?, 0, 0, ?)
        `,
			post.UserID,
			post.Content,
			post.Image_path,
			post.AllowComments,
			post.Location,
			post.GroupID,
			tags(),
		)

	} else if post.GroupID == -1 {

		result, err = db.Exec(`
            INSERT INTO posts (
                user_id,
                content,
                image_path,
                allow_comments,
                location,
                group_id,
                public,
                private,
                tags
            )
            VALUES (?, ?, ?, ?, ?, NULL, 0, 1, ?)
        `,
			post.UserID,
			post.Content,
			post.Image_path,
			post.AllowComments,
			post.Location,
			tags(),
		)

	} else {

		result, err = db.Exec(`
            INSERT INTO posts (
                user_id,
                content,
                image_path,
                allow_comments,
                location,
                group_id,
                public,
                private,
                tags
            )
            VALUES (?, ?, ?, ?, ?, NULL, 1, 0, ?)
        `,
			post.UserID,
			post.Content,
			post.Image_path,
			post.AllowComments,
			post.Location,
			tags(),
		)
	}

	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func GroupExists(db *sql.DB, groupID, userID int) error {
	if groupID == 0 || groupID == -1 {
		return nil
	}

	err := db.QueryRow(`
		SELECT 1 FROM user_posts_groups WHERE id = ? AND user_id = ?
	`, groupID, userID)

	return err.Err()
}

func GetHomePosts(db *sql.DB, userID, offset int) ([]models.Post, error) {
	var posts []models.Post

	seen := make(map[int]bool)
	tagsByPostIndex := make(map[int]string)

	appendPosts := func(rows *sql.Rows) error {
		defer rows.Close()

		for rows.Next() {
			var p models.Post
			var username sql.NullString
			var avatarPath sql.NullString
			var tags sql.NullString

			if err := rows.Scan(
				&p.Id,
				&p.UserId,
				&p.FirstName,
				&p.LastName,
				&username,
				&avatarPath,
				&p.Content,
				&p.ImagePath,
				&p.AllowComments,
				&p.Location,
				&p.CreatedAt,
				&p.GroupId,
				&tags,
				&p.ReactionValue,
				&p.LikeCount,
				&p.DisLikeCount,
				&p.CommentCount,
			); err != nil {
				return err
			}

			if username.Valid {
				p.Username = &username.String
			}

			if avatarPath.Valid {
				p.AvatarPath = avatarPath.String
			}

			if !seen[p.Id] {
				seen[p.Id] = true
				posts = append(posts, p)

				if tags.Valid {
					tagsByPostIndex[len(posts)-1] = tags.String
				}
			}
		}

		return rows.Err()
	}

	excludeClause := func() (string, []interface{}) {
		if len(posts) == 0 {
			return "", nil
		}

		placeholders := make([]string, len(posts))
		args := make([]interface{}, len(posts))

		for i, p := range posts {
			placeholders[i] = "?"
			args[i] = p.Id
		}

		return " AND p.id NOT IN (" + strings.Join(placeholders, ",") + ")", args
	}

	idsToPlaceholders := func(ids []int) (string, []interface{}) {
		if len(ids) == 0 {
			return "", nil
		}

		placeholders := make([]string, len(ids))
		args := make([]interface{}, len(ids))

		for i, id := range ids {
			placeholders[i] = "?"
			args[i] = id
		}

		return strings.Join(placeholders, ","), args
	}

	getFriendIDs := func() ([]int, error) {
		rows, err := db.Query(`
			SELECT uf1.target_id
			FROM user_followers uf1
			JOIN user_followers uf2
				ON uf1.target_id = uf2.follower_id
				AND uf1.follower_id = uf2.target_id
			WHERE uf1.follower_id = ?
				AND uf1.status = 1
				AND uf2.status = 1
		`, userID)

		if err != nil {
			return nil, err
		}

		defer rows.Close()

		var ids []int

		for rows.Next() {
			var id int

			if err := rows.Scan(&id); err != nil {
				return nil, err
			}

			ids = append(ids, id)
		}

		return ids, rows.Err()
	}

	getFollowingIDs := func(excludeIDs []int) ([]int, error) {
		exClause := "0"
		args := []interface{}{userID}

		if len(excludeIDs) > 0 {
			ph, exArgs := idsToPlaceholders(excludeIDs)
			exClause = ph
			args = append(args, exArgs...)
		}

		rows, err := db.Query(`
			SELECT target_id
			FROM user_followers
			WHERE follower_id = ?
				AND status = 1
				AND target_id NOT IN (`+exClause+`)
		`, args...)

		if err != nil {
			return nil, err
		}

		defer rows.Close()

		var ids []int

		for rows.Next() {
			var id int

			if err := rows.Scan(&id); err != nil {
				return nil, err
			}

			ids = append(ids, id)
		}

		return ids, rows.Err()
	}

	runGroupsQuery := func(unviewedOnly bool, limit int) error {
		if limit <= 0 {
			return nil
		}

		exClause, exArgs := excludeClause()

		viewClause := ""

		if unviewedOnly {
			viewClause = `
				AND NOT EXISTS (
					SELECT 1
					FROM post_views pv
					WHERE pv.post_id = p.id
						AND pv.user_id = ?
				)`
		}

		query := `
			SELECT
				p.id,
				p.user_id,
				u.first_name,
				u.last_name,
				u.username,
				pr.avatar_path,
				p.content,
				p.image_path,
				p.allow_comments,
				p.location,
				p.created_at,
				p.group_id,
				p.tags,
				COALESCE(prx.value, 0),
				p.like_count,
				p.dislike_count,
				p.comment_count
			FROM posts p
			JOIN user_posts_groups g
				ON p.group_id = g.id
			JOIN user u
				ON u.id = p.user_id
			LEFT JOIN profile pr
				ON pr.user_id = p.user_id
			LEFT JOIN post_reactions prx
				ON prx.post_id = p.id
				AND prx.user_id = ?
			WHERE (':' || g.users || ':') LIKE ('%:' || ? || ':%')` +
			exClause +
			viewClause + `
			ORDER BY p.created_at DESC
			LIMIT ?
		`

		args := []interface{}{userID, userID}

		args = append(args, exArgs...)

		if unviewedOnly {
			args = append(args, userID)
		}

		args = append(args, limit)

		rows, err := db.Query(query, args...)

		if err != nil {
			return err
		}

		return appendPosts(rows)
	}

	runUserPostsQuery := func(userIDs []int, includePrivate bool, unviewedOnly bool, limit int) error {
		if len(userIDs) == 0 || limit <= 0 {
			return nil
		}

		userPh, userArgs := idsToPlaceholders(userIDs)

		exClause, exArgs := excludeClause()

		visibilityClause := `
			AND p.public = 1
		`

		if includePrivate {
			visibilityClause = `
				AND (
					p.public = 1
					OR p.private = 1
				)
			`
		}

		viewClause := ""

		if unviewedOnly {
			viewClause = `
				AND NOT EXISTS (
					SELECT 1
					FROM post_views pv
					WHERE pv.post_id = p.id
						AND pv.user_id = ?
				)`
		}

		query := `
			SELECT
				p.id,
				p.user_id,
				u.first_name,
				u.last_name,
				u.username,
				pr.avatar_path,
				p.content,
				p.image_path,
				p.allow_comments,
				p.location,
				p.created_at,
				p.group_id,
				p.tags,
				COALESCE(prx.value, 0),
				p.like_count,
				p.dislike_count,
				p.comment_count
			FROM posts p
			JOIN user u
				ON u.id = p.user_id
			LEFT JOIN profile pr
				ON pr.user_id = p.user_id
			LEFT JOIN post_reactions prx
				ON prx.post_id = p.id
				AND prx.user_id = ?
			WHERE p.user_id IN (` + userPh + `)
				` + visibilityClause +
			exClause +
			viewClause + `
			ORDER BY p.created_at DESC
			LIMIT ?
		`

		var args []interface{}

		args = append(args, userID)
		args = append(args, userArgs...)
		args = append(args, exArgs...)

		if unviewedOnly {
			args = append(args, userID)
		}

		args = append(args, limit)

		rows, err := db.Query(query, args...)

		if err != nil {
			return err
		}

		return appendPosts(rows)
	}

	runRandomQuery := func(excludeUserIDs []int, unviewedOnly bool, limit int, withOffset bool) error {
		if limit <= 0 {
			return nil
		}

		exClause, exArgs := excludeClause()

		excludeUserClause := ""
		var excludeUserArgs []interface{}

		if len(excludeUserIDs) > 0 {
			ph, uargs := idsToPlaceholders(excludeUserIDs)

			excludeUserClause = " AND p.user_id NOT IN (" + ph + ")"
			excludeUserArgs = uargs
		}

		viewClause := ""

		if unviewedOnly {
			viewClause = `
				AND NOT EXISTS (
					SELECT 1
					FROM post_views pv
					WHERE pv.post_id = p.id
						AND pv.user_id = ?
				)`
		}

		offsetClause := ""

		if withOffset {
			offsetClause = " OFFSET ?"
		}

		query := `
			SELECT
				p.id,
				p.user_id,
				u.first_name,
				u.last_name,
				u.username,
				pr.avatar_path,
				p.content,
				p.image_path,
				p.allow_comments,
				p.location,
				p.created_at,
				p.group_id,
				p.tags,
				COALESCE(prx.value, 0),
				p.like_count,
				p.dislike_count,
				p.comment_count
			FROM posts p
			JOIN user u
				ON u.id = p.user_id
			LEFT JOIN profile pr
				ON pr.user_id = p.user_id
			LEFT JOIN post_reactions prx
				ON prx.post_id = p.id
				AND prx.user_id = ?
			WHERE p.public = 1
				AND p.user_id != ?` +
			excludeUserClause +
			exClause +
			viewClause + `
			ORDER BY p.created_at DESC
			LIMIT ?` +
			offsetClause

		args := []interface{}{userID, userID}

		args = append(args, excludeUserArgs...)
		args = append(args, exArgs...)

		if unviewedOnly {
			args = append(args, userID)
		}

		args = append(args, limit)

		if withOffset {
			args = append(args, offset)
		}

		rows, err := db.Query(query, args...)

		if err != nil {
			return err
		}

		return appendPosts(rows)
	}

	friendIDs, err := getFriendIDs()
	if err != nil {
		return nil, err
	}

	followingIDs, err := getFollowingIDs(friendIDs)
	if err != nil {
		return nil, err
	}

	if err := runGroupsQuery(true, 5); err != nil {
		return nil, err
	}

	if len(posts) < 9 {
		remaining := 9 - len(posts)

		if remaining > 4 {
			remaining = 4
		}

		if err := runUserPostsQuery(
			friendIDs,
			true,
			true,
			remaining,
		); err != nil {
			return nil, err
		}
	}

	if len(posts) < 12 {
		remaining := 12 - len(posts)

		if remaining > 3 {
			remaining = 3
		}

		if err := runUserPostsQuery(
			followingIDs,
			true,
			true,
			remaining,
		); err != nil {
			return nil, err
		}
	}

	if len(posts) < 13 {
		remaining := 13 - len(posts)

		excluded := append(
			append([]int{}, friendIDs...),
			followingIDs...,
		)

		if err := runRandomQuery(
			excluded,
			true,
			remaining,
			false,
		); err != nil {
			return nil, err
		}
	}

	if len(posts) < 13 {
		remaining := 13 - len(posts)

		if err := runGroupsQuery(false, remaining); err != nil {
			return nil, err
		}
	}

	if len(posts) < 13 {
		remaining := 13 - len(posts)

		if err := runUserPostsQuery(
			friendIDs,
			true,
			false,
			remaining,
		); err != nil {
			return nil, err
		}
	}

	if len(posts) < 13 {
		remaining := 13 - len(posts)

		if err := runUserPostsQuery(
			followingIDs,
			true,
			false,
			remaining,
		); err != nil {
			return nil, err
		}
	}

	if len(posts) < 13 {
		remaining := 13 - len(posts)

		excluded := append(
			append([]int{}, friendIDs...),
			followingIDs...,
		)

		if err := runRandomQuery(
			excluded,
			false,
			remaining,
			true,
		); err != nil {
			return nil, err
		}
	}

	if err := attachGroupOwnerNames(db, posts); err != nil {
		return nil, err
	}

	if err := attachTaggedPeople(db, posts, tagsByPostIndex); err != nil {
		return nil, err
	}

	return posts, nil
}

func attachGroupOwnerNames(db *sql.DB, posts []models.Post) error {
	groupIDSet := make(map[int]bool)
	for _, p := range posts {
		if p.GroupId != nil && *p.GroupId != 0 && *p.GroupId != -1 {
			groupIDSet[*p.GroupId] = true
		}
	}

	if len(groupIDSet) == 0 {
		return nil
	}

	placeholders := make([]string, 0, len(groupIDSet))
	args := make([]interface{}, 0, len(groupIDSet))
	for id := range groupIDSet {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}

	query := `
		SELECT g.id, u.first_name, u.last_name
		FROM user_posts_groups g
		JOIN user u ON u.id = g.user_id
		WHERE g.id IN (` + strings.Join(placeholders, ",") + `)
	`

	rows, err := db.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	ownerNames := make(map[int]string)
	for rows.Next() {
		var groupID int
		var firstName, lastName string
		if err := rows.Scan(&groupID, &firstName, &lastName); err != nil {
			return err
		}
		ownerNames[groupID] = firstName + " " + lastName
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for i := range posts {
		if posts[i].GroupId != nil {
			if name, ok := ownerNames[*posts[i].GroupId]; ok {
				posts[i].VisibilityUser = name
			}
		}
	}

	return nil
}

func attachTaggedPeople(db *sql.DB, posts []models.Post, tagsByPostIndex map[int]string) error {
	postTagIDs := make(map[int][]int)
	allIDSet := make(map[int]bool)

	for i := range posts {
		raw, ok := tagsByPostIndex[i]
		if !ok || raw == "" {
			continue
		}

		parts := strings.Split(raw, ":")
		var ids []int
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			id, err := strconv.Atoi(part)
			if err != nil {
				continue
			}
			ids = append(ids, id)
			allIDSet[id] = true
		}

		if len(ids) > 0 {
			postTagIDs[i] = ids
		}
	}

	if len(allIDSet) == 0 {
		return nil
	}

	placeholders := make([]string, 0, len(allIDSet))
	args := make([]interface{}, 0, len(allIDSet))
	for id := range allIDSet {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}

	query := `
		SELECT u.id, u.first_name, u.last_name, pr.avatar_path
		FROM user u
		LEFT JOIN profile pr ON pr.user_id = u.id
		WHERE u.id IN (` + strings.Join(placeholders, ",") + `)
	`

	rows, err := db.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	peopleByID := make(map[int]models.TaggedPerson)
	for rows.Next() {
		var id int
		var firstName, lastName string
		var avatarPath sql.NullString
		if err := rows.Scan(&id, &firstName, &lastName, &avatarPath); err != nil {
			return err
		}
		person := models.TaggedPerson{
			Id:        id,
			FirstName: firstName,
			LastName:  lastName,
		}
		if avatarPath.Valid {
			person.AvatarPath = avatarPath.String
		}
		peopleByID[id] = person
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for i, ids := range postTagIDs {
		var people []models.TaggedPerson
		for _, id := range ids {
			if person, ok := peopleByID[id]; ok {
				people = append(people, person)
			}
		}
		posts[i].TaggedPeople = people
	}

	return nil
}

func GetUserPosts(db *sql.DB, targetID, offset int) ([]models.Post, error) {
	var posts []models.Post

	tagsByPostIndex := make(map[int]string)

	rows, err := db.Query(`
		SELECT
			p.id,
			u.id,
			u.first_name,
			u.last_name,
			p.content,
			p.image_path,
			p.allow_comments,
			p.location,
			p.group_id,
			p.created_at,
			p.tags,
			p.like_count,
			p.dislike_count,
			p.comment_count,
			p.public,
			p.private
		FROM user u
		JOIN posts p
			ON p.user_id = u.id
		WHERE p.user_id = ?
		ORDER BY p.created_at DESC
		LIMIT 9
		OFFSET ?
	`, targetID, offset)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var avatar string

	err = db.QueryRow(`
		SELECT avatar_path
		FROM profile
		WHERE user_id = ?
	`, targetID).Scan(&avatar)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	for rows.Next() {
		var tags sql.NullString
		var p models.Post

		err := rows.Scan(
			&p.Id,
			&p.UserId,
			&p.FirstName,
			&p.LastName,
			&p.Content,
			&p.ImagePath,
			&p.AllowComments,
			&p.Location,
			&p.GroupId,
			&p.CreatedAt,
			&tags,
			&p.LikeCount,
			&p.DisLikeCount,
			&p.CommentCount,
			&p.Public,
			&p.Private,
		)

		if err != nil {
			return nil, err
		}

		p.AvatarPath = avatar

		if tags.Valid {
			tagsByPostIndex[len(posts)] = tags.String
		}

		if p.GroupId != nil && *p.GroupId > 0 {
			var groupName string

			err = db.QueryRow(`
				SELECT name
				FROM user_posts_groups
				WHERE id = ?
			`, *p.GroupId).Scan(&groupName)

			if err != nil && err != sql.ErrNoRows {
				return nil, err
			}

			p.GroupName = groupName
		}

		posts = append(posts, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := attachTaggedPeople(db, posts, tagsByPostIndex); err != nil {
		return nil, err
	}

	return posts, nil
}

func DeletePost(db *sql.DB, postID, userID int) error {
	result, err := db.Exec(`
		DELETE FROM posts
		WHERE id = ?
		AND user_id = ?
	`, postID, userID)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func ViewPost(db *sql.DB, postID, userID int) error {
	_, err := db.Exec(`
		INSERT INTO post_views (user_id, post_id)
		VALUES (?,?)
	`, postID, userID)
	return err
}

func SearchMembers(db *sql.DB, groupID int, search string) (map[int]models.UserRegistration, error) {
	search = strings.TrimSpace(search)

	rows, err := db.Query(`
		SELECT
			u.id,
			u.first_name,
			u.last_name,
			u.avatar
		FROM users u
		INNER JOIN groups_users gu
			ON gu.user_id = u.id
		WHERE gu.group_id = ?
		AND gu.status = 1
		AND (
			u.first_name LIKE ?
			OR u.last_name LIKE ?
			OR (u.first_name || ' ' || u.last_name) LIKE ?
		)
		ORDER BY u.first_name, u.last_name
		LIMIT 10
	`,
		groupID,
		"%"+search+"%",
		"%"+search+"%",
		"%"+search+"%",
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users := make(map[int]models.UserRegistration)

	for rows.Next() {
		var user models.UserRegistration

		err := rows.Scan(
			&user.ID,
			&user.FirstName,
			&user.LastName,
			&user.Avatar,
		)

		if err != nil {
			return nil, err
		}

		users[user.ID] = user
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
