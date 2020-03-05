// nolint gomnd
package sqlmore_test

import (
	"database/sql"
	"testing"

	"github.com/bingoohuang/sqlmore"
	"github.com/stretchr/testify/assert"
)

// person 结构体，对应到person表字段
type person struct {
	ID  string
	Age int
}

// personDao 定义对person表操作的所有方法
type personDao struct {
	CreateTable func()                         `sql:"create table person(id varchar(100), age int)"`
	Add         func(person)                   `sql:"insert into person(id, age) values(:id, :age)"`
	AddAll      func(...person)                `sql:"insert into person(id, age) values(:id, :age)"`
	Find        func(id string) person         `sql:"select id, age from person where id=:1"`
	ListAll     func() []person                `sql:"select id, age from person"`
	ListByID    func(string) []person          `sql:"select id, age from person where id=:1"`
	Delete      func(string) int               `sql:"delete from person where id=:1"`
	GetAge      func(string) struct{ Age int } `sql:"select age from person where id=:1"`

	GetAgeX func(string) (struct{ Age int }, error) `sql:"select age from person where xid=:1"`

	Err error // 添加这个字段，可以用来单独接收error信息
}

func TestDao(t *testing.T) {
	that := assert.New(t)

	// 生成DAO，自动创建dao结构体中的函数字段
	dao := &personDao{}
	that.Nil(sqlmore.CreateDao("sqlite3", openDB(t), dao))

	// 建表
	dao.CreateTable()
	// 插入
	dao.Add(person{"100", 100})
	// 查找
	that.Equal(person{"100", 100}, dao.Find("100"))
	// 刪除
	that.Equal(1, dao.Delete("100"))
	// 再找，找不到，返回零值
	that.Zero(dao.Find("100"))
	// 插入
	dao.Add(person{"200", 200})
	// 多值插入
	dao.AddAll(person{"300", 300}, person{"400", 400})
	// 列表
	that.Equal([]person{{"200", 200}, {"300", 300}, person{"400", 400}}, dao.ListAll())
	// 条件列表
	that.Equal([]person{{"200", 200}}, dao.ListByID("200"))
	// 匿名结构
	that.Equal(struct{ Age int }{Age: 200}, dao.GetAge("200"))

	that.Nil(dao.Err)
	ageX, err := dao.GetAgeX("200")
	that.Error(err)
	that.Zero(ageX)
	that.Error(dao.Err)

	// 条件列表
	that.Equal([]person{{"200", 200}}, dao.ListByID("200"))
	that.Nil(dao.Err) // 验证Err字段是否重置
}

func openDB(t *testing.T) *sql.DB {
	// 创建数据库连接池
	db, err := sql.Open("sqlite3", ":memory:")
	assert.Nil(t, err)
	return db
}
