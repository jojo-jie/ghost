package data

import (
	"fmt"
	"ghost/internal/conf"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo)

// Data .
type Data struct {
	// TODO wrapped database client
	Db  *gorm.DB
	Rdb *redis.Client
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	dsn := fmt.Sprintf("%s?charset=utf8mb4&parseTime=True&loc=Local", c.Database.Source)
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN: dsn,
	}), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}
	sqlDb, err := db.DB()
	if err != nil {
		return nil, nil, err
	}
	sqlDb.SetMaxIdleConns(int(c.Database.GetSetMaxIdleConns()))
	sqlDb.SetMaxOpenConns(int(c.Database.GetSetMaxOpenConns()))
	sqlDb.SetConnMaxLifetime(time.Duration(int(c.Database.GetSetConnMaxLifetime().GetSeconds())))

	// callback
	db.Callback().Query().After("gorm:query").Register("gorm:operating_time", updateTimeStampForCreateCallback)


	rdb := redis.NewClient(&redis.Options{
		Addr:     c.Redis.GetAddr(),
		Password: "",
		DB:       0,
	})

	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
		sqlDb.Close()
		rdb.Close()
	}
	return &Data{
		Db:  db,
		Rdb: rdb,
	}, cleanup, nil
}

func updateTimeStampForCreateCallback(db *gorm.DB)  {
	fmt.Println(db.Statement.Schema)
	if db.Statement.Schema!=nil {
		location, _ := time.LoadLocation("Asia/Shanghai")
		nowTime:=time.Now().In(location).Format("2006-01-02 15:04:05")
		field:=db.Statement.Schema.LookUpField("operating_time")
		if field != nil {
			db.Statement.Update("operating_time", nowTime)
		}
	}
}
