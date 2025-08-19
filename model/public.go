package model

import "ginWeb/utils/database"

// GetPermissionById 获取该用户所有权限
func GetPermissionById(userid int64) (permissions []string, err error) {
	db := database.Db
	permissions = make([]string, 0)
	// 获取所有权限
	rows, err := db.Raw(`
select permission_name
from permissions where id in (
select b.permission_id as id
from user_group a
         left join group_permission b on a.group_id = b.group_id
where user_id = ?
union all
select b.permission_id as id
from user_role a
         left join role_permission b on a.role_id = b.role_id
where user_id = ?);`, userid, userid).Rows()
	if err != nil {
		return
	}
	var p string
	for rows.Next() {
		err = rows.Scan(&p)
		if err != nil {
			return
		}
		permissions = append(permissions, p)
	}
	return
}

func GetPermissionByUUID(id string) (permissions []string, err error) {
	return
}

type BaseModel struct {
	Id        int64 `gorm:"primary_key;AUTO_INCREMENT"`
	CreatedAt int64 `gorm:"autoCreateTime:milli;index"`
	UpdatedAt int64 `gorm:"autoUpdateTime:milli"`
	Deleted   bool  `gorm:"type:boolean;default:false"`
}
