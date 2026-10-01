package routes

import (
	"database/sql"
	"log"
	"net/http"
	"path/filepath"
	"social/internal/app/api"
	"social/internal/app/mailer"
	"social/internal/app/otp"
	"sync"

	"golang.org/x/net/websocket"
)

func StartServer(db *sql.DB) *http.ServeMux {
	uploadsDir, err := filepath.Abs("uploads")
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()

	app := &api.App{
		DB: db,
		H: &api.Hub{
			Conn: make(map[int]*websocket.Conn),
			Mu:   sync.RWMutex{},
		},
		OTP: otp.NewStore(),
		Mail: mailer.New(mailer.Config{
			EmailPassword: "lmvm ugpc xvlo food",
			EmailAddress:  "almadhoonlinux@gmail.com",
		}),
	}

	//user
	mux.HandleFunc("GET /api/user", app.AuthMiddleware(app.GetUserData))
	mux.HandleFunc("POST /api/user", app.RegisterUser)
	mux.HandleFunc("PATCH /api/user", app.AuthMiddleware(app.UpdateUserInfo))
	mux.HandleFunc("DELETE /api/user", app.AuthMiddleware(app.DeleteAccount))
	
	//registration
	mux.HandleFunc("GET /api/registration/check", app.CheckAvailability)
	mux.HandleFunc("POST /api/email/code", app.SendEmailCode)
	mux.HandleFunc("POST /api/email/verify", app.VerifyEmailCode)

	//session
	mux.HandleFunc("POST /api/session", app.LoggingUser)
	mux.HandleFunc("GET /api/session", app.AuthMiddleware(app.AuthorizeSession))
	mux.HandleFunc("DELETE /api/session", app.DeleteSession)

	//profiles
	mux.HandleFunc("PATCH /api/profile/avatar", app.AuthMiddleware(app.UpdateUserAvatar))
	mux.HandleFunc("GET /api/profile/about", app.AuthMiddleware(app.GetUserAbout))
	mux.HandleFunc("PATCH /api/profile/about", app.AuthMiddleware(app.UpdateUserAbout))
	mux.HandleFunc("GET /api/profile", app.AuthMiddleware(app.GetUserProfile))

	// follow handler
	mux.HandleFunc("POST /api/profile/follow", app.AuthMiddleware(app.RequestFollow))
	mux.HandleFunc("DELETE /api/profile/follow", app.AuthMiddleware(app.CancelRequest))
	mux.HandleFunc("GET /api/profile/follow", app.AuthMiddleware(app.GetFollowers))
	mux.HandleFunc("GET /api/profile/following", app.AuthMiddleware(app.GetFollowing))
	mux.HandleFunc("DELETE /api/profile/followers", app.AuthMiddleware(app.RemoveFollower))
	mux.HandleFunc("/api/follow/accept", app.AuthMiddleware(app.AcceptFollowRequest))

	//folder handlers
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadsDir))))

	// posts
	mux.HandleFunc("GET /api/friends/", app.AuthMiddleware(app.GetFriends))
	mux.HandleFunc("POST /api/post", app.AuthMiddleware(app.AddPost))
	mux.HandleFunc("GET /api/posts", app.AuthMiddleware(app.GetHomePosts))
	mux.HandleFunc("POST /api/post/reaction", app.AuthMiddleware(app.PostReaction))
	mux.HandleFunc("GET /api/user/posts", app.AuthMiddleware(app.GetUserPosts))
	mux.HandleFunc("DELETE /api/post", app.AuthMiddleware(app.DeletePost))
	mux.HandleFunc("POST /api/posts/seen", app.AuthMiddleware(app.ViewPost))
	mux.HandleFunc("GET /api/post/single", app.AuthMiddleware(app.GetSinglePost))

	// post's groups
	mux.HandleFunc("GET /api/post/groups", app.AuthMiddleware(app.GetPostGroups))
	mux.HandleFunc("POST /api/post/groups", app.AuthMiddleware(app.AddPostGroup))
	mux.HandleFunc("DELETE /api/post/groups", app.AuthMiddleware(app.DeletePostGroup))
	mux.HandleFunc("PATCH /api/post/groups", app.AuthMiddleware(app.UpdatePostGroup))

	//comments
	mux.HandleFunc("POST /api/post/comment", app.AuthMiddleware(app.AddComment))
	mux.HandleFunc("GET /api/post/comment", app.AuthMiddleware(app.GetComments))
	mux.HandleFunc("DELETE /api/post/comment", app.AuthMiddleware(app.DeleteComment))
	mux.HandleFunc("POST /api/post/comment/vote", app.AuthMiddleware(app.VoteComment))

	// searches
	mux.HandleFunc("GET /api/profile/follows/search", app.AuthMiddleware(app.SearchFollows))
	mux.HandleFunc("GET /api/profile/following/search", app.AuthMiddleware(app.SearchFollowing))
	mux.HandleFunc("GET /api/location/search", app.SearchLocation)
	mux.HandleFunc("GET /api/groups/search", app.AuthMiddleware(app.SearchPrivateChats))
	mux.HandleFunc("GET /api/group/users", app.AuthMiddleware(app.GetGroupMembers))
	mux.HandleFunc("GET /api/search/groups", app.AuthMiddleware(app.GlobalSearchGroups))
	mux.HandleFunc("GET /api/search/users", app.AuthMiddleware(app.GlobalSearchUsers))
	mux.HandleFunc("GET /api/search/posts", app.AuthMiddleware(app.GlobalSearchPosts))
	mux.HandleFunc("GET /api/user/follow-followers", app.AuthMiddleware(app.GetFollowers_Following))

	// chats
	mux.HandleFunc("GET /api/groups", app.AuthMiddleware(app.GetGroups))
	mux.HandleFunc("POST /api/chats", app.AuthMiddleware(app.AddMessages))
	mux.HandleFunc("GET /api/chats", app.AuthMiddleware(app.GetMessages))

	//group chats
	mux.HandleFunc("GET /api/groups/invites/search", app.AuthMiddleware(app.SearchInvites))
	mux.HandleFunc("POST /api/groups", app.AuthMiddleware(app.MakeNewGroup))
	mux.HandleFunc("POST /api/groups/status", app.AuthMiddleware(app.AcceptInvite))
	mux.HandleFunc("GET /api/groups/discover", app.AuthMiddleware(app.DiscoverGroups))
	mux.HandleFunc("GET /api/group", app.AuthMiddleware(app.GetGroup))
	mux.HandleFunc("GET /api/group/search", app.AuthMiddleware(app.SearchMembers))
	mux.HandleFunc("POST /api/group/posts", app.AuthMiddleware(app.AddGroupPost))
	mux.HandleFunc("GET /api/group/post", app.AuthMiddleware(app.GetGroupPost))
	mux.HandleFunc("POST /api/group/post/reaction", app.AuthMiddleware(app.InsertGroupPostReaction))
	mux.HandleFunc("GET /api/group/posts", app.AuthMiddleware(app.GetGroupPosts))
	mux.HandleFunc("POST /api/group/invite", app.AuthMiddleware(app.InviteMember))
	mux.HandleFunc("POST /api/groups/request", app.AuthMiddleware(app.GroupRequest))
	mux.HandleFunc("GET /api/groups/requests", app.AuthMiddleware(app.GetGroupRequests))
	mux.HandleFunc("POST /api/groups/requests", app.AuthMiddleware(app.HandleGroupRequest))

	//preferences
	mux.HandleFunc("PATCH /api/user/preferences", app.AuthMiddleware(app.ChangePerferance))
	//group posts comments
	mux.HandleFunc("POST /api/group/post/comment", app.AuthMiddleware(app.AddGroupComment))
	mux.HandleFunc("GET /api/group/post/comment", app.AuthMiddleware(app.GetGroupComments))
	mux.HandleFunc("DELETE /api/group/post/comment", app.AuthMiddleware(app.DeleteGroupComment))
	mux.HandleFunc("POST /api/group/post/comment/vote", app.AuthMiddleware(app.VoteGroupComment))

	//group events
	mux.HandleFunc("POST /api/group/events", app.AuthMiddleware(app.AddGroupEvent))
	mux.HandleFunc("GET /api/group/events", app.AuthMiddleware(app.GetGroupEvents))
	mux.HandleFunc("POST /api/group/event/response", app.AuthMiddleware(app.RespondGroupEvent))
	mux.HandleFunc("GET /api/group/event/votes", app.AuthMiddleware(app.GetGroupEventVotes))

	//ws
	mux.Handle("/api/ws", app.WSAuthMiddleware(websocket.Handler(app.HandleWS)))
	mux.HandleFunc("/api/notifications", app.AuthMiddleware(app.GetNotification))
	mux.HandleFunc("GET /api/notifications/unread", app.AuthMiddleware(app.GetUnreadNotificationCount))
	mux.HandleFunc("POST /api/notifications/read", app.AuthMiddleware(app.MarkNotificationsRead))

	return mux
}
