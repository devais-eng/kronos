package db

import (
	"devais.it/kronos/internal/pkg/config"
	"devais.it/kronos/internal/pkg/db/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"io/ioutil"
	"path"
)

type SuiteBase struct {
	suite.Suite
	db     *gorm.DB
	dbConf config.DBConfig
}

func (s *SuiteBase) SetupSuite() {
	logrus.SetOutput(ioutil.Discard)

	assert := s.Require()

	err := config.InitGlobalEnvironment()
	assert.NoError(err)

	dbFilename := path.Join(s.T().TempDir(), "kronos-"+uuid.NewString()+"-test.db")
	s.dbConf = config.DefaultDBConfig()
	s.dbConf.URL = dbFilename

	err = OpenDB(&s.dbConf)
	assert.NoError(err, "Failed to open test Database")
	s.db = DB()
	s.T().Log("Database opened: ", dbFilename)
}

func (s *SuiteBase) TearDownSuite() {
	assert := s.Require()
	err := Close()
	assert.NoError(err, "Failed to close test Database")
	s.T().Log("Database closed")
}

func (s *SuiteBase) SetupTest() {
	assert := s.Require()
	for _, tableName := range models.GetTableNames() {
		tx := db.Exec("DELETE FROM " + tableName)
		assert.NoError(tx.Error, "Failed to clear table")
	}
	s.T().Log("Tables deleted")
	// Reconfigure SQLite, just to be sure
	//assert.NoError(
	//	configureSqlite(s.db, &s.dbConf),
	//	"Failed to configure SQLite",
	//)
}
