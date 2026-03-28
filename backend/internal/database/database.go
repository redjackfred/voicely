package database

import (
	"log"
	"time"

	"github.com/redjackfred/voicely/backend/models/"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 設為全域變數 (Singleton)，供其他 package 使用以避免連線洩漏 [1]
var DB *gorm.DB

func ConnectDB(dsn string) {
	var err error

	// 透過 postgres.Open() 建立連線，並在開發期間啟用 logger.Info 來檢視 GORM 生成的底層 SQL 語法 [2, 3]
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	// 務必處理連線錯誤，以避免系統在背景默默崩潰 [3]
	if err != nil {
		log.Fatalf("無法連線至 PostgreSQL 資料庫: %v", err)
	}

	// 取得底層的 *sql.DB 以進行進階設定
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("無法取得底層 SQL DB: %v", err)
	}

	// 實作連線池 (Connection Pool) 最佳化設定 [3, 4]
	sqlDB.SetMaxIdleConns(10)           // 設定最大閒置連線數 [3]
	sqlDB.SetMaxOpenConns(100)          // 設定最大開啟連線數，保持系統在高併發下的效能 [3]
	sqlDB.SetConnMaxLifetime(time.Hour) // 避免連線過期引發錯誤

	log.Println("PostgreSQL 資料庫連線與連線池設定成功！")

	// 執行 AutoMigrate，自動同步 Go Struct 與資料庫表結構
	err = DB.AutoMigrate(&models.User{}, &models.Store{}, &models.Call{})
	if err != nil {
		log.Fatalf("資料庫結構遷移失敗: %v", err)
	}

	log.Println("資料庫結構遷移完成！")
}
