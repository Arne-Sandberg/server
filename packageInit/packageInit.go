package packageInit

import (
	"os"

	"github.com/freecloudio/freecloud/auth"
	"github.com/freecloudio/freecloud/config"
	"github.com/freecloudio/freecloud/db"
	"github.com/freecloudio/freecloud/fs"
	"github.com/freecloudio/freecloud/utils"
	"github.com/freecloudio/freecloud/vfs"
	clog "gopkg.in/clog.v1"
)

var (
	filesystem fs.FilesystemProvider
	database   *db.GormDB
)

func Init() {
	// Init all components
	utils.SetupLogger()
	config.Init()

	filesystem, err := fs.NewDiskFilesystem(config.GetString("fs.base_directory"), config.GetInt("fs.tmp_data_expiry"))
	if err != nil {
		os.Exit(1)
	}

	database, err = db.NewGormDB(config.GetString("db.type"), config.GetString("db.host"), config.GetInt("db.port"), config.GetString("db.user"), config.GetString("db.password"), config.GetString("db.name"))
	if err != nil {
		clog.Fatal(0, "Database setup failed, bailing out!")
		os.Exit(1)
	}

	auth.Init(database, database, config.GetInt("auth.session_expiry"))
	err = vfs.InitVirtualFilesystem(filesystem, database)
	if err != nil {
		clog.Fatal(0, "Virtual Filesysem setup failed, bailing out!")
		os.Exit(1)
	}
}

func Deinit() {
	auth.Close()
	vfs.Close()
	filesystem.Close()
	database.Close()
	utils.CloseLogger()
}
