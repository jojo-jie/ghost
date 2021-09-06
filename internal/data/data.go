package data

import (
	"fmt"
	"ghost/internal/conf"
	"github.com/go-kratos/kratos/v2/log"
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
	Db *gorm.DB
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

	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{
		Db: db,
	}, cleanup, nil
}
