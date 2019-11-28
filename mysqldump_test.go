package sqlmore

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"bytes"
	"text/template"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestGetTablesOk(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	defer db.Close()

	rows := sqlmock.NewRows([]string{"Tables_in_Testdb"}).
		AddRow("Test_Table_1").
		AddRow("Test_Table_2")

	mock.ExpectQuery("^SHOW TABLES$").WillReturnRows(rows)

	m := &mySQLDump{db: db}
	result, err := m.getTables()

	if err != nil {
		t.Errorf("error was not expected while updating stats: %s", err)
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	expectedResult := []string{"Test_Table_1", "Test_Table_2"}

	if !reflect.DeepEqual(result, expectedResult) {
		t.Fatalf("expected %#v, got %#v", result, expectedResult)
	}
}

func TestGetTablesNil(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	defer db.Close()

	rows := sqlmock.NewRows([]string{"Tables_in_Testdb"}).
		AddRow("Test_Table_1").
		AddRow(nil).
		AddRow("Test_Table_3")

	mock.ExpectQuery("^SHOW TABLES$").WillReturnRows(rows)

	m := &mySQLDump{db: db}
	result, err := m.getTables()

	if err != nil {
		t.Errorf("error was not expected while updating stats: %s", err)
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	expectedResult := []string{"Test_Table_1", "", "Test_Table_3"}

	if !reflect.DeepEqual(result, expectedResult) {
		t.Fatalf("expected %#v, got %#v", expectedResult, result)
	}
}

func TestGetServerVersionOk(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	defer db.Close()

	rows := sqlmock.NewRows([]string{"Version()"}).
		AddRow("test_version")

	mock.ExpectQuery("^SELECT version()").WillReturnRows(rows)

	m := &mySQLDump{db: db}
	result, err := m.getServerVersion()

	if err != nil {
		t.Errorf("error was not expected while updating stats: %s", err)
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	expectedResult := "test_version"

	if !reflect.DeepEqual(result, expectedResult) {
		t.Fatalf("expected %#v, got %#v", expectedResult, result)
	}
}

func TestCreateTableSQLOk(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	defer db.Close()

	rows := sqlmock.NewRows([]string{"Table", "Create Table"}).
		AddRow("Test_Table", "CREATE TABLE 'Test_Table' (`id` int(11) NOT NULL AUTO_INCREMENT,"+
			"`s` char(60) DEFAULT NULL, "+
			"PRIMARY KEY (`id`)"+
			")ENGINE=InnoDB DEFAULT CHARSET=latin1")

	mock.ExpectQuery("^SHOW CREATE TABLE Test_Table$").WillReturnRows(rows)

	m := &mySQLDump{db: db}
	result, err := m.createTableSQL("Test_Table")

	if err != nil {
		t.Errorf("error was not expected while updating stats: %s", err)
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	expectedResult := "CREATE TABLE 'Test_Table' (`id` int(11) NOT NULL AUTO_INCREMENT," +
		"`s` char(60) DEFAULT NULL, PRIMARY KEY (`id`))" +
		"ENGINE=InnoDB DEFAULT CHARSET=latin1"

	if !reflect.DeepEqual(result, expectedResult) {
		t.Fatalf("expected %#v, got %#v", expectedResult, result)
	}
}

func TestCreateTableValuesOk(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "email", "name"}).
		AddRow(1, "test@test.de", "Test' Name 1").
		AddRow(2, "test2@test.de", "Test Name 2")

	mock.ExpectQuery("^SELECT (.+) FROM test$").WillReturnRows(rows)

	ds, _ := template.New("mysqldump_tableDataStart").Parse(tableDataTmplStart)
	de, _ := template.New("mysqldump_tableDataEnd").Parse(tableDataTmplEnd)

	var b bytes.Buffer

	m := &mySQLDump{db: db}
	err = m.createTableValues(ds, de, &b, "test")

	if err != nil {
		t.Errorf("error was not expected while updating stats: %s", err)
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	expectedResult := "\n--\n-- " +
		"Dumping data for table `test`\n--\n" +
		"LOCK TABLES `test` WRITE;\n" +
		"/*!40000 ALTER TABLE `test` DISABLE KEYS */;\n\n" +
		"INSERT INTO `test` VALUES ('1','test@test.de','Test'' Name 1')," +
		"('2','test2@test.de','Test Name 2');\n\n" +
		"/*!40000 ALTER TABLE `test` ENABLE KEYS */;\n" +
		"UNLOCK TABLES;\n\n"

	result := b.String()
	if !reflect.DeepEqual(result, expectedResult) {
		t.Fatalf("expected %#v, got %#v", expectedResult, result)
	}
}

func TestCreateTableValuesNil(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "email", "name"}).
		AddRow(1, nil, "Test Name 1").
		AddRow(2, "test2@test.de", "Test Name 2")

	mock.ExpectQuery("^SELECT (.+) FROM test$").WillReturnRows(rows)

	ds, _ := template.New("mysqldump_tableDataStart").Parse(tableDataTmplStart)
	de, _ := template.New("mysqldump_tableDataEnd").Parse(tableDataTmplEnd)

	var b bytes.Buffer

	m := &mySQLDump{db: db}
	err = m.createTableValues(ds, de, &b, "test")

	if err != nil {
		t.Errorf("error was not expected while updating stats: %s", err)
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	expectedResult := "\n--\n-- " +
		"Dumping data for table `test`\n" +
		"--\n" +
		"LOCK TABLES `test` WRITE;\n" +
		"/*!40000 ALTER TABLE `test` DISABLE KEYS */;\n\n" +
		"INSERT INTO `test` VALUES ('1',null,'Test Name 1')," +
		"('2','test2@test.de','Test Name 2');\n\n" +
		"/*!40000 ALTER TABLE `test` ENABLE KEYS */;\n" +
		"UNLOCK TABLES;\n\n"

	result := b.String()
	if !reflect.DeepEqual(result, expectedResult) {
		t.Fatalf("expected %#v, got %#v", expectedResult, result)
	}
}

// nolint funlen
func TestCreateTableOk(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	defer db.Close()

	createTableRows := sqlmock.NewRows([]string{"Table", "Create Table"}).
		AddRow("Test_Table", "CREATE TABLE 'Test_Table' ("+
			"`id` int(11) NOT NULL AUTO_INCREMENT,"+
			"`s` char(60) DEFAULT NULL, "+
			"PRIMARY KEY (`id`))ENGINE=InnoDB DEFAULT CHARSET=latin1")

	createTableValueRows := sqlmock.NewRows([]string{"id", "email", "name"}).
		AddRow(1, nil, "Test Name 1").
		AddRow(2, "test2@test.de", "Test Name 2")

	mock.ExpectQuery("^SHOW CREATE TABLE Test_Table$").WillReturnRows(createTableRows)
	mock.ExpectQuery("^SELECT (.+) FROM Test_Table$").WillReturnRows(createTableValueRows)

	ct, _ := template.New("mysqldump_createTable").Parse(createTableTmpl)
	ds, _ := template.New("mysqldump_tableDataStart").Parse(tableDataTmplStart)
	de, _ := template.New("mysqldump_tableDataEnd").Parse(tableDataTmplEnd)

	var b bytes.Buffer

	m := &mySQLDump{db: db}
	err = m.createTable(ct, ds, de, &b, "Test_Table")

	if err != nil {
		t.Errorf("error was not expected while updating stats: %s", err)
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	result := b.String()

	expectedResult := "\n" +
		"--\n" +
		"-- Table structure for table `Test_Table`\n" +
		"--\n" +
		"DROP TABLE IF EXISTS `Test_Table`;\n" +
		"/*!40101 SET @saved_cs_client     = @@character_set_client */;\n" +
		"/*!40101 SET character_set_client = utf8 */;\n" +
		"CREATE TABLE 'Test_Table' (" +
		"`id` int(11) NOT NULL AUTO_INCREMENT," +
		"`s` char(60) DEFAULT NULL, " +
		"PRIMARY KEY (`id`))ENGINE=InnoDB DEFAULT CHARSET=latin1;\n" +
		"/*!40101 SET character_set_client = @saved_cs_client */;\n" +
		"--\n" +
		"-- Dumping data for table `Test_Table`\n" +
		"--\n" +
		"LOCK TABLES `Test_Table` WRITE;\n" +
		"/*!40000 ALTER TABLE `Test_Table` DISABLE KEYS */;\n" +
		"\n" +
		"INSERT INTO `Test_Table` VALUES ('1',null,'Test Name 1'),('2','test2@test.de','Test Name 2');\n" +
		"\n" +
		"/*!40000 ALTER TABLE `Test_Table` ENABLE KEYS */;\n" +
		"UNLOCK TABLES;\n" +
		"\n"
	if !reflect.DeepEqual(result, expectedResult) {
		t.Fatalf("expected %#v, got %#v", expectedResult, result)
	}
}

// nolint funlen
func TestDumpOk(t *testing.T) {
	tmpFile := "/tmp/test_format.sql"
	os.Remove(tmpFile)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	defer db.Close()

	showTablesRows := sqlmock.NewRows([]string{"Tables_in_Testdb"}).
		AddRow("Test_Table")

	serverVersionRows := sqlmock.NewRows([]string{"Version()"}).
		AddRow("test_version")

	createTableRows := sqlmock.NewRows([]string{"Table", "Create Table"}).
		AddRow("Test_Table", "CREATE TABLE 'Test_Table' ("+
			"`id` int(11) NOT NULL AUTO_INCREMENT,"+
			"`email` char(60) DEFAULT NULL,"+
			" `name` char(60),"+
			" PRIMARY KEY (`id`))ENGINE=InnoDB DEFAULT CHARSET=latin1")

	createTableValueRows := sqlmock.NewRows([]string{"id", "email", "name"}).
		AddRow(1, nil, "Test Name 1").
		AddRow(2, "test2@test.de", "Test Name 2")

	mock.ExpectQuery("^SELECT version()").WillReturnRows(serverVersionRows)
	mock.ExpectQuery("^SHOW TABLES$").WillReturnRows(showTablesRows)
	mock.ExpectQuery("^SHOW CREATE TABLE Test_Table$").WillReturnRows(createTableRows)
	mock.ExpectQuery("^SELECT (.+) FROM Test_Table$").WillReturnRows(createTableValueRows)

	var b bytes.Buffer
	err = MySQLDump(db, &b)
	if err != nil {
		t.Errorf("error was not expected while dumping the database: %s", err)
	}

	f := b.String()
	result := strings.Replace(strings.Split(f, "-- Dump completed")[0], "`", "\\", -1)

	expected := `-- Go SQL Dump ` + version + `
--
-- ------------------------------------------------------
-- Server version	test_version
/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table \Test_Table\
--
DROP TABLE IF EXISTS \Test_Table\;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE 'Test_Table' (` +
		`\id\ int(11) NOT NULL AUTO_INCREMENT,` +
		`\email\ char(60) DEFAULT NULL, ` +
		`\name\ char(60), PRIMARY KEY (\id\))` +
		`ENGINE=InnoDB DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;
--
-- Dumping data for table \Test_Table\
--
LOCK TABLES \Test_Table\ WRITE;
/*!40000 ALTER TABLE \Test_Table\ DISABLE KEYS */;

INSERT INTO \Test_Table\ VALUES ('1',null,'Test Name 1'),('2','test2@test.de','Test Name 2');

/*!40000 ALTER TABLE \Test_Table\ ENABLE KEYS */;
UNLOCK TABLES;

`

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected %#v, got %#v", expected, result)
	}
}
