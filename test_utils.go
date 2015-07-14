package main

import (
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/drone/config"
	"github.com/stretchr/testify/suite"
	"gopkg.in/mgo.v2"
)

// InitAllSuite has a InitSuite method, which will run before the
// tests in the suite are run.
type InitAllSuite interface {
	InitSuite()
}

// TearDownAllSuiteTests has a TearDownSuiteTests method, which will run after
// all the tests in the suite have been run.
type CleanAllSuite interface {
	CleanSuite()
}

// Base suite struct for embedding into a custom test cases.
type BaseSuite struct {
	suite.Suite
}

// Base suite struct for embedding into a custom test cases with database.
type BaseSuiteWithDB struct {
	BaseSuite

	session *mgo.Session
	db      *mgo.Database
}

// The InitSuite method will be run by testify once, at the very
// start of the testing suite, before any tests and Setup* methods
// are run.
func (suite *BaseSuiteWithDB) InitSuite() {
	fmt.Println("Database initialization")
	config.SetPrefix("TEST_")
	config.Parse("test.conf")
	suite.session = getSession()
	suite.db = suite.session.DB(*MongoDatabase)
	// remove old test database if exists
	err := suite.db.DropDatabase()
	if err != nil {
		suite.T().Error(err.Error())
	}
}

// The CleanSuite method will be run by testify once, at the very
// end of the testing suite, after all tests and TearDown* methods
// have been run.
func (suite *BaseSuiteWithDB) CleanSuite() {
	// remove test database
	err := suite.db.DropDatabase()
	if err != nil {
		suite.T().Error(err.Error())
	}
	suite.session.Close()
	fmt.Println("Database was removed successfully")
}

// The SetupSuite method will be run by testify once, after the InitSuite
// method, before any tests are run.
// Override this method in custom test cases.
func (suite *BaseSuite) SetupSuite() {}

// The TearDownSuite method will be run by testify once, before CleanSuite
// method, after all tests have been run.
// Override this method in custom test cases.
func (suite *BaseSuite) TearDownSuite() {}

// The SetupTest method will be run before every test in the suite.
// Override this method in custom test cases.
func (suite *BaseSuite) SetupTest() {}

// The TearDownTest method will be run after every test in the suite.
// Override this method in custom test cases.
func (suite *BaseSuite) TearDownTest() {}

// Run takes a testing suite and runs all of the tests attached to it.
func Run(t *testing.T, testingSuite suite.TestingSuite) {
	if initAllSuite, ok := testingSuite.(InitAllSuite); ok {
		initAllSuite.InitSuite()
	}

	defer func() {
		if cleanAllSuite, ok := testingSuite.(CleanAllSuite); ok {
			cleanAllSuite.CleanSuite()
		}
	}()

	suite.Run(t, testingSuite)
}

// RandomMD5 returns random hex string looks like MD5.
func RandomMD5() string {
	b := make([]byte, 32)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
