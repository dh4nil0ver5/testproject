package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	calculator "testproject/task"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

// task 1
type Student struct {
	StudentID int    `json:"student_id"`
	Name      string `json:"name"`
	// Add other fields as needed based on your mahasiswa table
}

type Students struct {
	gorm.Model
	Name         string
	StudentID    int
	StartingYear int
	Courses      []Courses
}

type Courses struct {
	gorm.Model
	CoursesID      int
	StudentID      int
	Name           string
	NumberOfCredit int
	Lecture        []Lecture
}
type Lecture struct {
	gorm.Model
	LecturerID int
	name       string
}

var (
	db            *sql.DB
	mutex         sync.Mutex
	randomNumbers []int
	srv           *http.Server
	running       = false
	wg            sync.WaitGroup
	err           error

	student = []Students{
		{Name: "Dani", StudentID: 10200112, StartingYear: 2012},
		{Name: "Deni", StudentID: 10200111, StartingYear: 2011},
		{Name: "Dina", StudentID: 10200113, StartingYear: 2011},
	}
	courses = []Courses{
		{CoursesID: 201, Name: "Perangkat Lunak I", StudentID: 10200112, NumberOfCredit: 0},
		{CoursesID: 202, Name: "Sistem Informasi Dasar I", StudentID: 10200112, NumberOfCredit: 0},
		{CoursesID: 203, Name: "Math Linear Grafik I", StudentID: 10200112, NumberOfCredit: 0},
		{CoursesID: 204, Name: "Math Linear Grafik I", StudentID: 10200111, NumberOfCredit: 0},
		{CoursesID: 205, Name: "Sistem Informasi Dasar I", StudentID: 10200111, NumberOfCredit: 0},
		{CoursesID: 206, Name: "Math Linear Grafik I", StudentID: 10200113, NumberOfCredit: 0},
		{CoursesID: 207, Name: "Sistem Informasi Dasar I", StudentID: 10200113, NumberOfCredit: 0},
	}
	lecture = []Lecture{
		{LecturerID: 4001, name: "Pak Budi"},
		{LecturerID: 4002, name: "Pak Darma"},
		{LecturerID: 4003, name: "Bu Tari"},
	}
)

// task 2
func populateDataStudent(w http.ResponseWriter, r *http.Request, db *calculator.Database) {
	// Iterate through student data and insert into the database using SQL queries
	for _, student := range student {
		err := db.InsertStudent(student.Name, fmt.Sprint(student.StudentID), student.StartingYear) // Convert StudentID to string
		if err != nil {
			panic(err) // Handle error appropriately
		}
	}
	fmt.Fprintf(w, "Data Student created !")
}

// task 2
func populateDataLecture(w http.ResponseWriter, r *http.Request, db *calculator.Database) {
	// Iterate through student data and insert into the database using SQL queries
	for _, lecture := range lecture {
		err := db.InsertLecture(lecture.name, fmt.Sprint(lecture.LecturerID)) // Convert StudentID to string
		if err != nil {
			panic(err) // Handle error appropriately
		}
	}
	fmt.Fprintf(w, "Data Lecture created !")
}

// task 2
func populateDataCourses(w http.ResponseWriter, r *http.Request, db *calculator.Database) {
	// Iterate through student data and insert into the database using SQL queries
	for _, courses := range courses {
		// (name, courses_id, student_id, number_of_credit)
		err := db.InsertCourse(courses.Name, courses.CoursesID, courses.StudentID, courses.NumberOfCredit) // Convert StudentID to string
		if err != nil {
			panic(err) // Handle error appropriately
		}
	}
	fmt.Fprintf(w, "Data Courses created !")
}

func handleError(w http.ResponseWriter, err error, statusCode int) {
	fmt.Fprintf(w, "Error: %v", err)
	w.WriteHeader(statusCode)
}

// task 3
func handleDataMahasiswa(w http.ResponseWriter, r *http.Request, db *calculator.Database) {
	// Extract nim parameter from the URL query string
	nim := r.URL.Query().Get("nim")
	if nim == "" {
		fmt.Fprintf(w, "Error: Missing nim parameter in query string.")
		return
	}

	db, err := calculator.Test("localhost", 5432, "postgres", "admin", "testapi")
	if err != nil {
		fmt.Fprintf(w, "Error: Database connection failed.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// defer rows.Close()
	var wg sync.WaitGroup
	wg.Add(2) // Wait for two goroutines

	// Goroutine to retrieve student data
	var student Student
	go func() {
		defer wg.Done()
		rows, err := db.GetStudentByNIM(nim)
		if err != nil {
			handleError(w, err, http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		if rows.Next() {
			err := rows.Scan(&student.StudentID, &student.Name)
			if err != nil {
				handleError(w, err, http.StatusInternalServerError)
				return
			}
		} else {
			fmt.Fprintf(w, "No student found with student_id=%s", nim)
			w.WriteHeader(http.StatusNotFound) // Set not found status code
			return
		}
	}()

	// Goroutine to retrieve courses
	var courseSlice []map[string]interface{}
	go func() {
		defer wg.Done()
		courses, err := db.GetCoursesByNim(nim)
		if err != nil {
			handleError(w, err, http.StatusInternalServerError)
			return
		}
		defer courses.Close()

		for courses.Next() {
			var studentID int
			var name string
			var numberOfCredit int
			err := courses.Scan(&studentID, &name, &numberOfCredit)
			if err != nil {
				handleError(w, err, http.StatusInternalServerError)
				return
			}
			course := map[string]interface{}{
				"student_id":       studentID,
				"name":             name,
				"number_of_credit": numberOfCredit,
			}
			courseSlice = append(courseSlice, course)
		}

		if len(courseSlice) == 0 {
			courseSlice = []map[string]interface{}{} // Empty slice for no courses
		}
	}()

	// Wait for both goroutines to finish
	wg.Wait()

	if len(courseSlice) == 0 {
		courseSlice = []map[string]interface{}{} // Empty slice for no courses
	}

	responseData := map[string]interface{}{
		"student_id": student.StudentID,
		"name":       student.Name,
		"courses":    courseSlice,
	}

	jsonData, err := json.Marshal(responseData)
	if err != nil {
		fmt.Fprintf(w, "Error encoding student data to JSON: %v", err)
		w.WriteHeader(http.StatusInternalServerError) // Set internal server error status code
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

// Task 4 -> createfile for looping random to roman text
func startConcurrency() {
	wg.Add(1) // Add a waiting goroutine
	go func() {
		defer wg.Done()
		rand.Seed(time.Now().UnixNano())
		for {
			if !running {
				return
			}
			num := rand.Intn(1000000000)
			mutex.Lock()
			fmt.Println("Random number:", num)

			f, err := os.OpenFile("last_random_number.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Println("Error opening file for append:", err)
				return
			}
			defer f.Close()

			_, err = f.Seek(0, io.SeekEnd)
			if err != nil {
				fmt.Println("Error seeking to end of file:", err)
				return
			}

			data := strconv.Itoa(num) + "\n"
			_, err = f.Write([]byte(data))
			if err != nil {
				fmt.Println("Error writing to file:", err)
				return
			}

			mutex.Unlock()
			time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
		}
	}()
	running = true // Set running flag to true
}

func stopConcurrency() {
	running = false
	wg.Wait()

	mutex.Lock()
	defer mutex.Unlock()
}

func main() {

	db, err := calculator.Test("localhost", 5432, "postgres", "admin", "testapi")
	if err != nil {
		fmt.Printf("db have failure connection, please check again !")
		return
	}
	defer db.Close()

	http.HandleFunc("/task5/startconcurreny", func(w http.ResponseWriter, r *http.Request) {
		startConcurrency()
		w.Write([]byte("Concurrent process started\n"))
	})

	http.HandleFunc("/task5/stopconcurreny", func(w http.ResponseWriter, r *http.Request) {
		stopConcurrency()
		w.Write([]byte("Concurrent process stopped\n"))
	})

	//
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	//task 4
	http.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		stopConcurrency() // Stop the concurrent process before exiting
		fmt.Println("Shutting down server...")

		// Create a Server object with a 5-second timeout
		srv := &http.Server{Addr: ":8080", Handler: nil} // Use your handler here

		// Gracefully shut down the server with 5-second timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			fmt.Println("Server shutdown error:", err)
		}

		os.Exit(0)
	})

	http.HandleFunc("/task2/preparingdata", func(w http.ResponseWriter, r *http.Request) {
		populateDataStudent(w, r, db)
		populateDataCourses(w, r, db)
		populateDataLecture(w, r, db)
	})
	//
	http.HandleFunc("/task3/datamahasiswa", func(w http.ResponseWriter, r *http.Request) {
		handleDataMahasiswa(w, r, db)
	})

	http.HandleFunc("/task6/readingromantxt", func(w http.ResponseWriter, r *http.Request) {
		// Define the file path
		folderPath := "C:\\Users\\Administrator\\Downloads\\"
		fileName := "roman(1).txt"
		filePath := folderPath + fileName
		content, err := readTextFilesFromFolder(filePath)
		if err != nil {
			fmt.Fprintf(w, "Error reading file: %v", err)
			return
		}

		fmt.Fprintf(w, "Content of %s:\n%s\n", fileName, content)

		// fmt.Println("Content:", content)
		number, err := romanToInteger(content)
		if err != nil {
			fmt.Println("Error converting Roman numeral:", err)
			return
		}
		fmt.Println("Integer value from Roman numeral:", number)
	})

	fmt.Println("Server listening on port 8080")
	http.ListenAndServe(":8080", nil)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Println("Server listening error:", err)
	}
}
