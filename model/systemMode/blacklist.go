package systemMode

import (
	"errors"
	"ginWeb/model"
	"ginWeb/utils/database"
)

type UserBlacklist struct {
	model.BaseModel `gorm:"embedded"`
	UserUuid        string `gorm:"size:36;index;not null;"`
	TargetUuid      string `gorm:"size:36;index;not null;"`
	Description     string `gorm:"size:255;default:''"`
}

func (u UserBlacklist) Add() error {
	if u.UserUuid == "" || u.TargetUuid == "" {
		return errors.New("blank uuid")
	}
	tx := database.Db.Begin()
	defer tx.Commit()
	var exist int64
	resp := tx.Table("user").Where("user_uuid = ? and target_uuid = ?", u.UserUuid, u.TargetUuid).Count(&exist)
	if resp.Error != nil || exist == 0 {
		return errors.New("target not exsit")
	}
	resp = tx.Table("user_blacklist").Create(u)

	return resp.Error
}

// 标记删除
func (u UserBlacklist) Delete() error {
	if u.UserUuid == "" || u.TargetUuid == "" {
		return errors.New("blank uuid")
	}
	tx := database.Db.Begin()
	defer tx.Commit()
	resp := tx.Table("user_blacklist").Where("user_uuid = ? AND target_uuid = ?", u.UserUuid, u.TargetUuid).Update("deleted", true)
	return resp.Error
}

// 是否在黑名单中
func ExistInList(uuid string) (bool, error) {
	var count int64
	resp := database.Db.Table("user_blacklist").Where("target_uuid = ?", uuid).Count(&count)
	return count >= 1, resp.Error
}

type DumpData struct {
	Username    string `json:"username"`
	Description string `json:"description"`
}

// 获取user的所有uuid黑名单
func DumpUserBlacklist(uuid string) (data []DumpData, err error) {
	if uuid == "" {
		return nil, errors.New("not exist")
	}
	resp := database.Db.Exec("select b.username, a.description from user_blacklist as a inner join user as b on a.target_uuid = b.uuid").Scan(data)
	if resp.Error != nil {
		return nil, resp.Error
	}
	return
}
