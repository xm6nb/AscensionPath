package model

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	Username  string    `gorm:"type:varchar(50);uniqueIndex;not null"`
	Password  string    `gorm:"type:varchar(100);not null"` // 存储加密后的密码
	Email     string    `gorm:"type:varchar(100);uniqueIndex"`
	Status    int       `gorm:"type:tinyint(1);default:0"`       // 状态(0:禁用 1:正常)
	Score     float64   `gorm:"type:int;default:0"`              // 新增分数字段
	Role      string    `gorm:"type:varchar(20);default:'user'"` // 新增身份字段
	LastLogin time.Time `gorm:"default:null"`                    // 最后登录时间
	CreatedAt time.Time `gorm:"autoCreateTime"`                  // 创建时间
	UpdatedAt time.Time `gorm:"autoUpdateTime"`                  // 更新时间
}

// CreateUser 创建用户
func CreateUser(user *User) error {
	return DB.Create(user).Error
}

// GetUserByID 通过ID获取用户
func GetUserByID(id uint) (*User, error) {
	var user User
	result := DB.First(&user, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// GetUserByUsername 通过用户名获取用户
func GetUserByUsername(username string) (*User, error) {
	var user User
	result := DB.Where("username = ?", username).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// UpdateUser 更新用户信息（指定字段）
func UpdateUser(id uint, updates map[string]interface{}) error {
	return DB.Model(&User{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// DeleteUser 删除用户
func DeleteUser(id uint) error {
	return DB.Delete(&User{}, id).Error
}

// UpdateUserScore 更新用户积分
func UpdateUserScore(id uint, score float64) error {
	return DB.Model(&User{}).
		Where("id =?", id).
		Update("score", score).Error
}

// UpdatePassword 更新密码
func UpdatePassword(id uint, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return DB.Model(&User{}).Where("id = ?", id).Update("password", string(hashedPassword)).Error
}

// UserExists 检查用户是否存在
func UserExists(username, email string) (bool, error) {
	var count int64
	err := DB.Model(&User{}).
		Where("username = ? OR email = ?", username, email).
		Count(&count).Error
	return count > 0, err
}

// UpdateUserRole 更新用户身份
func UpdateUserRole(id uint, newRole string) error {
	err := DB.Model(&User{}).Where("id = ?", id).Update("role", newRole).Error
	return err
}

// GetAllUsers 获取所有用户（分页版）
func GetAllUsers(page, pageSize int) ([]User, error) {
	var users []User
	result := DB.Scopes(Paginate(page, pageSize)).Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

// 分页通用方法
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}
		if pageSize <= 0 {
			pageSize = 10 // 默认每页10条
		}
		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

// GetUserCount 获取所有用户的总数
func GetUserCount() (int64, error) {
	var count int64
	err := DB.Model(&User{}).Count(&count).Error
	return count, err
}

// SearchUsers 搜索用户
func SearchUsers(username, email string, status int, role string, page, pageSize int) ([]User, error) {
	var users []User
	db := DB.Model(&User{})

	if username != "" {
		db = db.Where("username LIKE ?", "%"+username+"%")
	}
	if email != "" {
		db = db.Where("email LIKE ?", "%"+email+"%")
	}
	if status >= 0 { // 使用-1表示不筛选状态
		db = db.Where("status = ?", status)
	}
	if role != "" {
		db = db.Where("role = ?", role)
	}

	result := db.Scopes(Paginate(page, pageSize)).Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

// SearchUsersCount 统计符合搜索条件的用户总数
func SearchUsersCount(username, email string, status int, role string) (int64, error) {
	var count int64

	query := DB.Model(&User{})
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if email != "" {
		query = query.Where("email LIKE ?", "%"+email+"%")
	}
	if status != -1 {
		query = query.Where("status = ?", status)
	}
	if role != "" {
		query = query.Where("role = ?", role)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
