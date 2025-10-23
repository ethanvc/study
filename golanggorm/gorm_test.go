package golanggorm

import (
	"fmt"
	"testing"

	"github.com/VividCortex/mysqlerr"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	drivermysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 定义数据模型
type User struct {
	Id       int64  `gorm:"primaryKey"`                    // 主键字段
	UserName string `gorm:"type:varchar(255);uniqueIndex"` // 唯一索引字段
}

/*
docker run --name test-mysql \
  -e MYSQL_ALLOW_EMPTY_PASSWORD=yes \
  -e MYSQL_ROOT_HOST='%' \
  -e MYSQL_DATABASE=test_db \
  -p 3306:3306 \
  -d mysql:8.0 \
  --skip-log-bin
*/

func TestDuplicateInsert(t *testing.T) {
	// MySQL连接配置 (按需修改)
	user := "root"
	pass := ""
	host := "localhost"
	port := "3306"
	dbName := "test_db"

	// 创建数据库连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, pass, host, port, dbName)

	// 连接MySQL
	db, err := gorm.Open(drivermysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // 显示SQL日志
	})
	require.NoError(t, err)

	err = db.AutoMigrate(&User{})
	require.NoError(t, err)

	// 插入测试数据
	userInst := &User{
		UserName: "john",
	}

	result := db.Create(userInst)
	require.Error(t, result.Error)
	mysqlErr, _ := result.Error.(*mysql.MySQLError)
	require.Equal(t, uint16(mysqlerr.ER_DUP_ENTRY), mysqlErr.Number)
}

func TestSkipInitializeWithVersion(t *testing.T) {
	dsn := "root:@tcp(127.0.0.1:3306)/test_db?charset=utf8mb4"
	dialConf := drivermysql.Config{
		// if fetch mysql version from server instance.
		SkipInitializeWithVersion: false,
		DSN:                       dsn,
	}
	db, err := gorm.Open(drivermysql.New(dialConf), &gorm.Config{})
	require.NoError(t, err)
	_ = db
}

func TestGormPlugin(t *testing.T) {
	// https://gorm.io/docs/write_plugins.html
	// https://levelup.gitconnected.com/how-to-implement-a-gorm-plugin-in-golang-f359501950de
	dsn := "root:@tcp(127.0.0.1:3306)/test_db?charset=utf8mb4"
	dialConf := drivermysql.Config{
		// if fetch mysql version from server instance.
		SkipInitializeWithVersion: false,
		DSN:                       dsn,
	}
	db, err := gorm.Open(drivermysql.New(dialConf), &gorm.Config{})
	require.NoError(t, err)
	db.Use(&GormPlugin{})
	cb := &GormCallback{}
	callback := db.Callback().Query()
	callback.After("*").Register("go_report", cb.AfterQuery)
	var users *UserTab
	err = db.Table("user_tab").Model(users).Take(&users).Error
	require.NoError(t, err)
}

type UserTab struct {
	Id       int64 `gorm:"primaryKey"`
	UserName string
	Extra    datatypes.JSON
}

type ExtraContent struct {
	Finished bool
}

// Name set the table name
func (tab *UserTab) TableName() string {
	return "user_tab"
}

type GormCallback struct {
}

func (cb *GormCallback) AfterQuery(db *gorm.DB) {
	sqlStr := db.Statement.SQL.String()
	_ = sqlStr
	fmt.Println(db)
}

type GormPlugin struct {
}

func (p *GormPlugin) Name() string {
	return "gorm_plugin"
}

func (p *GormPlugin) Initialize(db *gorm.DB) error {
	return nil
}

// https://github.com/go-gorm/datatypes

// INSERT INTO `user_tab` (`user_name`,`extra`) VALUES ('test',CAST('{"Finished": false}' AS JSON))
// set to ture or false
// UPDATE `user_tab` SET `extra`=JSON_SET(`extra`,?,CAST(? AS JSON)) WHERE id=?
// UPDATE `user_tab` SET `extra`=JSON_SET(`extra`,?,?) WHERE id=? args are $.Finished true 1
func TestGormJsonBool(t *testing.T) {
	db := createDb()
	require.NotNil(t, db)
	u := &UserTab{
		UserName: "test",
		Extra:    []byte(`{"Finished": false}`),
	}
	require.NoError(t, db.Create(u).Error)
	// direct use true or false will set value to 0 or 1
	// when use datatypes.NewJSONType, the second arg become string(true)
	// add a breakpoint aat sql.go:Tx.ExecContext to check the parameter.
	err := db.Model(&UserTab{}).Where("id=?", 1).UpdateColumn(
		"extra",
		datatypes.JSONSet("extra").Set("Finished", datatypes.NewJSONType(true)),
	).Error
	require.NoError(t, err)
}

func createDb() *gorm.DB {
	dsn := "root:@tcp(127.0.0.1:3306)/test_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(drivermysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	err = db.AutoMigrate(&UserTab{})
	if err != nil {
		panic(err)
	}
	return db
}

func TestTranslateError(t *testing.T) {
	// MySQL连接配置 (按需修改)
	user := "root"
	pass := ""
	host := "localhost"
	port := "3306"
	dbName := "test_db"

	// 创建数据库连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, pass, host, port, dbName)

	// 连接MySQL
	db, err := gorm.Open(drivermysql.Open(dsn), &gorm.Config{
		Logger:         logger.Default.LogMode(logger.Info), // 显示SQL日志
		TranslateError: true,
	})
	require.NoError(t, err)

	err = db.AutoMigrate(&User{})
	require.NoError(t, err)

	// 插入测试数据
	userInst := &User{
		UserName: "john",
	}

	result := db.Create(userInst)
	require.ErrorIs(t, result.Error, gorm.ErrDuplicatedKey)
}
