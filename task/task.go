package calculator

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Database struct {
	db *sql.DB
}

func Test(host string, port int, user string, password string, dbName string) (*Database, error) {
	dataSourceName := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbName)

	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}

	if err := db.PingContext(context.Background()); err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}

func (db *Database) Close() error {
	return db.db.Close()
}

func (db Database) GetStudentByNIM(nim string) (*sql.Rows, error) {
	// Implement the logic to retrieve student by NIM using db.db
	rows, err := db.ExecQuery("SELECT student_id, name FROM students WHERE student_id = $1", nim)

	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (db Database) GetCoursesByNim(nim string) (*sql.Rows, error) {
	// Implement logic to retrieve courses using nim
	rows, err := db.ExecQuery("select a.student_id as student_id, c.name as name, c.number_of_credit as number_of_credit FROM testapi.public.students a inner join testapi.public.courses c on c.student_id=a.student_id WHERE c.student_id = $1", nim)

	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (db *Database) InsertStudent(name string, studentID string, startingYear int) error {
	stmt, err := db.db.Prepare("INSERT INTO students (name, student_id, starting_year) VALUES ($1, $2, $3)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Use Exec to execute the prepared statement with student data
	_, err = stmt.Exec(name, studentID, startingYear)
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) InsertLecture(name string, lecture_id string) error {
	stmt, err := db.db.Prepare("INSERT INTO lectures (name, lecturer_id) VALUES ($1, $2)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Use Exec to execute the prepared statement with student data
	_, err = stmt.Exec(name, lecture_id)
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) InsertCourse(name string, courses_id int, student_id int, number_of_credit int) error {
	stmt, err := db.db.Prepare("INSERT INTO Courses (name, courses_id, student_id, number_of_credit) VALUES ($1, $2, $3, $4)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Use Exec to execute the prepared statement with student data
	_, err = stmt.Exec(name, courses_id, student_id, number_of_credit)
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) ExecQuery(query string, args ...interface{}) (*sql.Rows, error) {

	rows, err := db.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}
