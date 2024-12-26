package test

import (
	"campus2/pkg/global"
	"fmt"
	"testing"
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	gorm.Model
	Name     string     `gorm:"size:32;not null"`
	Age      int        `gorm:"not null"`
	Email    string     `gorm:"size:128;uniqueIndex"`
	Profile  Profile    `gorm:"foreignKey:UserID"`
	Articles []Article  `gorm:"many2many:user_articles;"`
	Address  *Address   `gorm:"foreignKey:UserID"`
	Orders   []Order    `gorm:"foreignKey:UserID"`
	Role     string     `gorm:"type:enum('admin','user');default:'user'"`
	Tags     []UserTag  `gorm:"polymorphic:Tagged;"`
	DeleteAt *time.Time `gorm:"index"`
}

// Profile 用户档案
type Profile struct {
	ID        uint   `gorm:"primarykey"`
	UserID    uint   `gorm:"uniqueIndex"`
	Biography string `gorm:"type:text"`
}

// Article 文章
type Article struct {
	gorm.Model
	Title    string    `gorm:"size:128;not null"`
	Content  string    `gorm:"type:text"`
	Authors  []User    `gorm:"many2many:user_articles;"`
	Comments []Comment `gorm:"foreignKey:ArticleID"`
}

// Comment 评论
type Comment struct {
	gorm.Model
	ArticleID uint   `gorm:"index"`
	Content   string `gorm:"type:text"`
}

// Address 地址
type Address struct {
	gorm.Model
	UserID  uint   `gorm:"index"`
	Street  string `gorm:"size:256"`
	City    string `gorm:"size:64"`
	Country string `gorm:"size:64"`
}

// Order 订单
type Order struct {
	gorm.Model
	UserID      uint    `gorm:"index"`
	OrderNumber string  `gorm:"size:32;uniqueIndex"`
	Amount      float64 `gorm:"type:decimal(10,2)"`
	Status      string  `gorm:"type:enum('pending','paid','shipped','delivered');default:'pending'"`
}

// UserTag 用户标签（多态关联示例）
type UserTag struct {
	ID         uint   `gorm:"primarykey"`
	Name       string `gorm:"size:32"`
	TaggedID   uint
	TaggedType string
	CreatedAt  time.Time
}

func init() {
	// 自动迁移数据库结构
	db := global.GVA_DB
	err := db.AutoMigrate(
		&User{},
		&Profile{},
		&Article{},
		&Comment{},
		&Address{},
		&Order{},
		&UserTag{},
	)
	if err != nil {
		panic("数据库迁移失败: " + err.Error())
	}
}

// TestCRUD 测试基本的CRUD操作
func TestCRUD(t *testing.T) {
	db := global.GVA_DB

	// 创建用户
	user := User{
		Name:  "测试用户",
		Age:   25,
		Email: "test@example.com",
		Profile: Profile{
			Biography: "这是一个测试用户",
		},
	}

	// Create
	result := db.Create(&user)
	if result.Error != nil {
		t.Fatalf("创建用户失败: %v", result.Error)
	}
	t.Logf("创建用户成功，ID: %d", user.ID)

	// Read
	var foundUser User
	result = db.Preload("Profile").First(&foundUser, user.ID)
	if result.Error != nil {
		t.Fatalf("查询用户失败: %v", result.Error)
	}
	t.Logf("查询用户成功: %+v", foundUser)

	// Update
	result = db.Model(&foundUser).Update("Age", 26)
	if result.Error != nil {
		t.Fatalf("更新用户失败: %v", result.Error)
	}
	t.Log("更新用户年龄成功")

	// Delete
	result = db.Delete(&foundUser)
	if result.Error != nil {
		t.Fatalf("删除用户失败: %v", result.Error)
	}
	t.Log("删除用户成功")
}

// TestAssociations 测试关联关系
func TestAssociations(t *testing.T) {
	db := global.GVA_DB

	// 创建测试数据
	user := User{
		Name:  "关联测试用户",
		Age:   30,
		Email: "assoc@example.com",
		Profile: Profile{
			Biography: "用于测试关联关系",
		},
		Address: &Address{
			Street:  "测试街道",
			City:    "测试城市",
			Country: "测试国家",
		},
	}

	// 创建用户及其关联数据
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("创建用户及关联数据失败: %v", err)
	}

	// 创建文章
	article := Article{
		Title:   "测试文章",
		Content: "这是一篇测试文章",
	}
	if err := db.Create(&article).Error; err != nil {
		t.Fatalf("创建文章失败: %v", err)
	}

	// 添加多对多关联
	if err := db.Model(&user).Association("Articles").Append(&article); err != nil {
		t.Fatalf("添加文章关联失败: %v", err)
	}

	// 测试预加载
	var fullUser User
	if err := db.Preload("Profile").
		Preload("Address").
		Preload("Articles").
		First(&fullUser, user.ID).Error; err != nil {
		t.Fatalf("预加载查询失败: %v", err)
	}

	t.Logf("完整用户数据: %+v", fullUser)
	t.Logf("用户地址: %+v", fullUser.Address)
	t.Logf("用户文章数量: %d", len(fullUser.Articles))
}

// TestRedisTransactions 测试事务
func TestRedisTransactions(t *testing.T) {
	db := global.GVA_DB

	// 开始事务
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("开启事务失败: %v", tx.Error)
	}

	// 在事务中执行操作
	user := User{
		Name:  "事务测试用户",
		Age:   35,
		Email: "tx@example.com",
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		t.Fatalf("在事务中创建用户失败: %v", err)
	}

	// 故意制造一个错误
	if err := tx.Model(&user).Update("Age", "invalid").Error; err != nil {
		tx.Rollback()
		t.Log("事务回滚成功")
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("提交事务失败: %v", err)
	}
}

// TestQueries 测试高级查询
func TestQueries(t *testing.T) {
	db := global.GVA_DB

	// 准备测试数据
	users := []User{
		{Name: "用户1", Age: 20, Email: "user1@example.com"},
		{Name: "用户2", Age: 25, Email: "user2@example.com"},
		{Name: "用户3", Age: 30, Email: "user3@example.com"},
	}
	db.Create(&users)

	// 测试条件查询
	var result []User
	if err := db.Where("age > ?", 22).Find(&result).Error; err != nil {
		t.Fatalf("条件查询失败: %v", err)
	}
	t.Logf("年龄大于22的用户数量: %d", len(result))

	// 测试分页
	var pagedUsers []User
	if err := db.Offset(1).Limit(2).Find(&pagedUsers).Error; err != nil {
		t.Fatalf("分页查询失败: %v", err)
	}
	t.Logf("分页查询结果数量: %d", len(pagedUsers))

	// 测试排序
	var orderedUsers []User
	if err := db.Order("age desc").Find(&orderedUsers).Error; err != nil {
		t.Fatalf("排序查询失败: %v", err)
	}
	t.Log("按年龄降序排序成功")

	// 测试分组和聚合
	type Result struct {
		Age   int
		Count int64
	}
	var aggregateResult []Result
	if err := db.Model(&User{}).
		Select("age, count(*) as count").
		Group("age").
		Find(&aggregateResult).Error; err != nil {
		t.Fatalf("分组聚合查询失败: %v", err)
	}
	t.Logf("年龄分组统计结果: %+v", aggregateResult)
}

// TestHooks 测试钩子函数
func TestHooks(t *testing.T) {
	// 为User添加钩子方法
	user := &User{
		Name:  "钩子测试用户",
		Age:   40,
		Email: "hooks@example.com",
	}

	db := global.GVA_DB
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("创建带钩子的用户失败: %v", err)
	}

	if err := db.Delete(user).Error; err != nil {
		t.Fatalf("删除带钩子的用户失败: %v", err)
	}
}

// 钩子方法
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// 创建前的处理
	fmt.Println("BeforeCreate")
	if u.Age < 0 {
		return gorm.ErrInvalidField
	}
	return nil
}

func (u *User) AfterCreate(tx *gorm.DB) error {
	// 创建后的处理
	fmt.Println("AfterCreate")
	return nil
}

func (u *User) BeforeDelete(tx *gorm.DB) error {
	// 删除前的处理
	fmt.Println("BeforeDelete")
	return nil
}

func (u *User) AfterDelete(tx *gorm.DB) error {
	// 删除后的处理
	fmt.Println("AfterDelete")
	return nil
}
