package packageInit

import (
	"github.com/freecloudio/freecloud/auth"
	"github.com/freecloudio/freecloud/config"
	"github.com/freecloudio/freecloud/db"
	"github.com/freecloudio/freecloud/fs"
	"github.com/freecloudio/freecloud/vfs"
	log "gopkg.in/clog.v1"
)

var (
	filesystem fs.FilesystemProvider
	database   *db.GormDB
)

// Init initialize all modules/components
func Init() {
	config.Init()

	var err error
	filesystem, err = fs.NewDiskFilesystem(config.GetString("fs.base_directory"), config.GetInt("fs.tmp_data_expiry"))
	if err != nil {
		log.Fatal(0, "Filesyste setup failed, bailing out!")
	}

	database, err = db.NewGormDB(config.GetString("db.type"), config.GetString("db.host"), config.GetInt("db.port"), config.GetString("db.user"), config.GetString("db.password"), config.GetString("db.name"))
	if err != nil {
		log.Fatal(0, "Database setup failed, bailing out!")
	}

	auth.Init(database, database, config.GetInt("auth.session_expiry"))
	err = vfs.InitVirtualFilesystem(filesystem, database)
	if err != nil {
		log.Fatal(0, "Virtual Filesysem setup failed, bailing out!")
	}
}

// Deinit deinitializes all modules/components
func Deinit() {
	auth.Close()
	vfs.Close()
	filesystem.Close()
	database.Close()
}
