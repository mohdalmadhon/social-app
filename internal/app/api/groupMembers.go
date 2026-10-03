package api

import (
	"fmt"
	"log"

	"social/database/groups"
	"social/database/preferences"
	"social/database/users"
	"social/internal/models"
)

const (
	memberSkipped   = "skipped"
	memberRequested = "requested"
)

func (app *App) groupRelation(inviterID, targetID int) (isFriend bool, allowed bool, err error) {
	isFriend, err = users.IsFriend(app.DB, inviterID, targetID)

	if err != nil || isFriend {
		return isFriend, isFriend, err
	}

	following, err := users.IsFollowing(app.DB, inviterID, targetID)

	if err != nil || following {
		return false, following, err
	}

	follower, err := users.IsFollowing(app.DB, targetID, inviterID)

	return false, follower, err
}

func (app *App) addOrInvite(inviterID, targetID, groupID int, groupTitle string) (string, error) {
	if targetID <= 0 || targetID == inviterID {
		return memberSkipped, nil
	}

	alreadyIN, err := groups.UserIN(app.DB, groupID, targetID)

	if err != nil {
		return memberSkipped, err
	}

	if alreadyIN {
		return memberSkipped, nil
	}

	_, allowed, err := app.groupRelation(inviterID, targetID)

	if err != nil {
		return memberSkipped, err
	}

	if !allowed {
		return memberSkipped, nil
	}

	canInvite, err := preferences.CanInviteToGroup(app.DB, inviterID, targetID)

	if err != nil {
		return memberSkipped, err
	}

	if !canInvite {
		return memberSkipped, nil
	}

	if err := groups.AddPendingMember(app.DB, groupID, targetID, inviterID); err != nil {
		return memberSkipped, err
	}

	actor := inviterID
	gid := groupID

	delivered := app.notify(inviterID, models.NewNotification{
		UserID:            targetID,
		Message:           fmt.Sprintf("%s invited you to join the group \"%s\"", app.actorName(inviterID), groupTitle),
		GroupInviteUserID: &actor,
		GroupID:           &gid,
	})

	if !delivered {
		if err := groups.RemoveMember(app.DB, groupID, targetID); err != nil {
			log.Println("could not roll back pending member:", err)
		}

		return memberSkipped, fmt.Errorf("could not deliver invite")
	}

	return memberRequested, nil
}
