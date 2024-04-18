package services

import (
	"fmt"
	"imgu2/db"
)

type group struct{}

var Group = group{}

// Get the user group object for the user.
// If user is nil, the group for guest user is returned.
func (*group) GetUserGroup(user *db.User) (*db.Group, error) {
	var groupId int

	// guest
	if user == nil {
		var err error
		groupId, err = Setting.DefaultGroupGuest()
		if err != nil {
			return nil, err
		}
	} else {
		groupId = user.GroupId
	}

	g, err := db.GroupFindById(groupId)
	if err != nil {
		return nil, err
	}

	if g == nil {
		if user == nil {
			return nil, fmt.Errorf("guest group id does not exist: %d", groupId)
		}
		return nil, fmt.Errorf("group id does not exist: id=%d user=%d name=%s", groupId, user.Id, user.Username)
	}

	return g, nil
}

func (*group) FindAll() ([]db.Group, error) {
	return db.GroupFindAll()
}

// returns (nil, nil) if the group id does not exist
func (*group) FindById(id int) (*db.Group, error) {
	return db.GroupFindById(id)
}

// create a new user with random generated name
func (*group) Create() (int, error) {
	return db.GroupCreate()
}

func (*group) CountUsers(id int) (int, error) {
	return db.GroupCountUsers(id)
}

func (*group) Delete(id int) error {
	return db.GroupDelete(id)
}

func (*group) Edit(id int, name string, allow_upload bool, max_file_size, upload_per_minute, upload_per_hour, upload_per_day, upload_per_month, total_uploads, max_retention_seconds int) error {
	return db.GroupEdit(id, name, allow_upload, max_file_size, upload_per_minute, upload_per_hour, upload_per_day, upload_per_month, total_uploads, max_retention_seconds)
}
