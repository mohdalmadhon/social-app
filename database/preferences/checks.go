package preferences

import (
	"database/sql"
	"time"

	"social/database/users"
	"social/internal/models"
)

type Visibility struct {
	Email                  bool   `json:"email"`
	DOB                    bool   `json:"dob"`
	AdditionalInfo         bool   `json:"additionalInfo"`
	AdditionalInfoAudience string `json:"additionalInfoAudience"`
}

func ApplyProfileVisibility(db *sql.DB, viewerID, ownerID int, data *models.UserData) (Visibility, error) {
	prefs, err := GetPreferences(db, ownerID)
	if err == sql.ErrNoRows {
		prefs = Preferences{Chat: "following-followers", Email: "none", DOB: "none", AdditionalInfo: "any", GroupInvite: "following"}
	} else if err != nil {
		return Visibility{}, err
	}

	vis := Visibility{
		Email:                  prefs.Email == "any",
		DOB:                    prefs.DOB == "any",
		AdditionalInfoAudience: prefs.AdditionalInfo,
	}

	switch prefs.AdditionalInfo {
	case "any":
		vis.AdditionalInfo = true
	case "followers":
		vis.AdditionalInfo, err = users.IsFollowing(db, viewerID, ownerID)
	case "friends":
		vis.AdditionalInfo, err = users.IsFriend(db, viewerID, ownerID)
	}
	if err != nil {
		return Visibility{}, err
	}

	if !vis.Email {
		data.UserInfo.Email = ""
	}

	if !vis.DOB {
		data.UserInfo.DOB = time.Time{}
	}

	if !vis.AdditionalInfo {
		bio := data.About.Bio
		data.About = models.UserAbout{Bio: bio}
	}

	return vis, nil
}
