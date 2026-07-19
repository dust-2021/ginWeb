package systemMode

import (
	"context"
	"fmt"
	"time"

	"errors"
	"ginWeb/model"
	reCache "ginWeb/service/cache"
	"ginWeb/utils/database"
)

type UserBlacklist struct {
	model.BaseModel `gorm:"embedded"`
	UserUuid        string `gorm:"size:36;index;not null;"`
	TargetUuid      string `gorm:"size:36;index;not null;"`
	Type            int    `gorm:"default:0;index;not null;comment:'0: user, 1: ip, 2: device'"`
	Description     string `gorm:"size:255;default:''"`
}

func (u UserBlacklist) Add() error {
	if u.UserUuid == "" || u.TargetUuid == "" {
		return errors.New("blank uuid")
	}
	tx := database.Db.Begin()
	defer tx.Commit()
	var exist int64
	resp := tx.Table("user").Where("user_uuid = ? and target_uuid = ? and type = ?", u.UserUuid, u.TargetUuid, u.Type).Count(&exist)
	if resp.Error != nil || exist == 0 {
		return errors.New("target not exsit")
	}
	resp = tx.Table("user_blacklist").Create(u)
	reCache.Del("blacklist", u.UserUuid)
	// 重置缓存
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
	// 重置缓存
	reCache.Del("blacklist", u.UserUuid)
	return resp.Error
}

// 是否在黑名单中
func ExistInList(userUuid string, targetUuid string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	resp := database.Rdb.SIsMember(ctx, fmt.Sprintf("::blacklist::%s", userUuid), targetUuid)
	if resp.Err() == nil {
		return resp.Val(), nil
	}
	// 缓存中不存在，查询数据库
	var ids []UserBlacklist
	database.Db.Table("user_blacklist").Where("target_uuid = ? and deleted = false", targetUuid).Find(&ids)

	var exist bool = false
	for _, v := range ids {
		if v.UserUuid == userUuid {
			exist = true
		}
		database.Rdb.SAdd(ctx, fmt.Sprintf("::blacklist::%s", v.UserUuid), v.TargetUuid)
	}
	return exist, nil
}
