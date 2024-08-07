package controller

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	//"fmt"
	"log"
	"net/http"

	"new/new-go/model"

	"new/new-go/config"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

const jwtKey = "0273080777"

func generateToken(userID, patientID int, username string, DoctorID int) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID
	claims["patient_id"] = patientID
	claims["doctor_id"] = DoctorID
	claims["username"] = username
	claims["exp"] = time.Now().Add(time.Hour * 1).Unix()

	tokenString, err := token.SignedString([]byte(jwtKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func Signin(w http.ResponseWriter, r *http.Request) {
	var requestUser model.User
	var response model.Response

	// Decode JSON request body
	if err := json.NewDecoder(r.Body).Decode(&requestUser); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Open a database connection
	db := config.Connect()
	defer db.Close()

	// Query the database
	var databaseRetrieved model.User
	// Query the database with a JOIN operation
	err := db.QueryRow(`
    SELECT 
        u.user_id, 
        COALESCE(p.patient_id, 0) AS patient_id, 
        COALESCE(d.doctor_id, 0) AS doctor_id, 
        u.username, 
        u.email, 
        u.user_type
    FROM 
        users u
    LEFT JOIN 
        patients p ON u.user_id = p.user_id
    LEFT JOIN 
        doctors d ON u.user_id = d.user_id
    WHERE 
        u.email = ? AND u.password = ?

`, requestUser.Email, requestUser.Password).Scan(
		&databaseRetrieved.UserID,
		&databaseRetrieved.PatientID,
		&databaseRetrieved.DoctorID,
		&databaseRetrieved.Username,
		&databaseRetrieved.Email,
		&databaseRetrieved.UserType,
	)

	if err == sql.ErrNoRows {
		// No matching user found
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	} else if err != nil {
		// Other database error
		log.Print(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Generate JWT token
	token, err := generateToken(databaseRetrieved.UserID, databaseRetrieved.PatientID, databaseRetrieved.Username, databaseRetrieved.DoctorID)
	if err != nil {
		http.Error(w, "Failed to create JWT token", http.StatusInternalServerError)
		return
	}

	// Prepare response
	response.Status = http.StatusOK
	response.Message = "Sign-in successful"
	response.Data = []model.User{databaseRetrieved}
	response.Token = token

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Encode and send response
	json.NewEncoder(w).Encode(response)
}

func Signup(w http.ResponseWriter, r *http.Request) {
	var newUser model.User
	var response model.Response

	db := config.Connect()
	defer db.Close()

	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", newUser.Email).Scan(&count)
	if err != nil {
		log.Print(err)
		http.Error(w, "Error checking email availability", http.StatusInternalServerError)
		return
	}

	if count > 0 {
		http.Error(w, "Email already in use", http.StatusBadRequest)
		return
	}
	result, err := db.Exec("INSERT INTO users(username, password, email, user_type) VALUES(?, ?, ?, ?)",
		newUser.Username, newUser.Password, newUser.Email, newUser.UserType)
	if err != nil {
		log.Print(err)
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	userID, err := result.LastInsertId()
	if err != nil {
		log.Print(err)
		http.Error(w, "Error getting user ID", http.StatusInternalServerError)
		return
	}

	response.Status = http.StatusOK
	response.Message = "User registered successfully"

	if newUser.UserType == "Doctor" {
		_, err = db.Exec("INSERT INTO doctors(user_id, full_name) VALUES(?, ?)",
			userID, newUser.Username)
		if err != nil {
			log.Print(err)
			http.Error(w, "Error creating doctor", http.StatusInternalServerError)
			return
		}
	}
	if newUser.UserType == "Patient" {
		_, err = db.Exec("INSERT INTO patients(user_id, full_name) VALUES(?, ?)",
			userID, newUser.Username)
		if err != nil {
			log.Print(err)
			http.Error(w, "Error creating doctor", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(response)
}
func SetSchedule(w http.ResponseWriter, r *http.Request) {
	var request model.ScheduleRequest
	var response model.Response

	db := config.Connect()
	defer db.Close()

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var existingSlotCount int
	err := db.QueryRow("SELECT COUNT(*) FROM schedules WHERE doctor_id = ? AND date = ? AND hour = ?", request.DoctorID, request.Date, request.Hour).Scan(&existingSlotCount)
	if err != nil {
		log.Print(err)
		http.Error(w, "Error checking for existing slot", http.StatusInternalServerError)
		return
	}

	if existingSlotCount > 0 {
		http.Error(w, "Slot with the same date and hour already exists", http.StatusBadRequest)
		return
	}

	var doctorExists bool
	var doctorFullName string

	err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM doctors WHERE doctor_id = ?)", request.DoctorID).Scan(&doctorExists)
	if err != nil {
		log.Print(err)
		http.Error(w, "Error checking doctor existence", http.StatusInternalServerError)
		return
	}

	if !doctorExists {
		http.Error(w, "Doctor not found", http.StatusNotFound)
		return
	}

	err = db.QueryRow("SELECT full_name FROM doctors WHERE doctor_id = ?", request.DoctorID).Scan(&doctorFullName)
	if err != nil {
		log.Print(err)
		http.Error(w, "Error fetching doctor full name", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("INSERT INTO schedules(doctor_id, full_name, date, hour, status) VALUES(?, ?, ?, ?, 'available')", request.DoctorID, doctorFullName, request.Date, request.Hour)

	if err != nil {
		log.Print(err)
		http.Error(w, "Error creating schedule", http.StatusInternalServerError)
		return
	}

	response.Status = http.StatusOK
	response.Message = "Schedule created successfully"

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(response)
}

func GetAvailableDoctors(w http.ResponseWriter, r *http.Request) {
	db := config.Connect()
	defer db.Close()

	rows, err := db.Query("SELECT DISTINCT full_name FROM doctors")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var doctors []string
	for rows.Next() {
		var doctorFullName string
		err := rows.Scan(&doctorFullName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		doctors = append(doctors, doctorFullName)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(doctors)
}
func GetAvailableSlotsForDoctor(w http.ResponseWriter, r *http.Request) {
	DoctorName := mux.Vars(r)["full_name"]
	log.Printf("Doctor ID: %s", DoctorName)

	db := config.Connect()
	defer db.Close()

	query := "SELECT schedule_id, date, hour, status, full_name, doctor_id FROM schedules WHERE full_name = ? AND status = 'available'"
	rows, err := db.Query(query, DoctorName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var slots []model.ScheduleRequest
	for rows.Next() {
		var slot model.ScheduleRequest
		err := rows.Scan(&slot.ScheduleID, &slot.Date, &slot.Hour, &slot.Status, &slot.DoctorFullName, &slot.DoctorID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		slots = append(slots, slot)
	}
	if err = rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(slots)
}

func ChooseSlot(w http.ResponseWriter, r *http.Request) {
	var request model.ScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db := config.Connect()
	defer db.Close()

	var doctorExists bool
	err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM doctors WHERE doctor_id = ?)", request.DoctorID).Scan(&doctorExists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var slotExists bool
	err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM schedules WHERE schedule_id = ?)", request.ScheduleID).Scan(&slotExists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !doctorExists || !slotExists {
		http.Error(w, "Doctor or slot does not exist", http.StatusBadRequest)
		return
	}

	var doctorFullName string
	err = db.QueryRow("SELECT full_name FROM doctors WHERE doctor_id = ?", request.DoctorID).Scan(&doctorFullName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var slotStatus string
	err = db.QueryRow("SELECT status FROM schedules WHERE schedule_id = ?", request.ScheduleID).Scan(&slotStatus)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if slotStatus == "booked" {
		http.Error(w, "Slot is already booked", http.StatusBadRequest)
		return
	}

	query := "UPDATE schedules SET status = ?, patient_id = ?, full_name = ? WHERE schedule_id = ?"
	_, err = db.Exec(query, "booked", request.PatientID, doctorFullName, request.ScheduleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := model.Response{
		Status:  http.StatusOK,
		Message: "Slot chosen successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func CreateAppointment(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ScheduleID int `json:"schedule_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db := config.Connect()
	defer db.Close()

	// Retrieve doctor ID and patient ID from the schedule
	var doctorID, patientID int
	err := db.QueryRow("SELECT doctor_id, patient_id FROM schedules WHERE schedule_id = ?", request.ScheduleID).Scan(&doctorID, &patientID)
	if err != nil {
		http.Error(w, "Error retrieving doctor and patient information", http.StatusInternalServerError)
		return
	}

	// Check for existing appointment
	var existingAppointmentCount int
	err = db.QueryRow("SELECT COUNT(*) FROM appointments WHERE doctor_id = ? AND schedule_id = ?", doctorID, request.ScheduleID).Scan(&existingAppointmentCount)
	if err != nil {
		http.Error(w, "Error checking for existing appointments", http.StatusInternalServerError)
		return
	}

	if existingAppointmentCount > 0 {
		http.Error(w, "Appointment with the same schedule already exists", http.StatusBadRequest)
		return
	}

	// Retrieve schedule details
	var schedule model.ScheduleRequest
	err = db.QueryRow("SELECT date, hour, patient_id, schedule_id, full_name FROM schedules WHERE schedule_id = ?", request.ScheduleID).Scan(&schedule.Date, &schedule.Hour, &schedule.PatientID, &schedule.ScheduleID, &schedule.DoctorFullName)
	if err != nil {
		http.Error(w, "Error retrieving slot data", http.StatusInternalServerError)
		return
	}

	// Insert appointment into the database
	result, err := db.Exec("INSERT INTO appointments(doctor_id, appointment_date, appointment_hour, patient_id, status, schedule_id , doctor_name) VALUES (?, ?, ?, ?, 'booked', ?, ?)", doctorID, schedule.Date, schedule.Hour, schedule.PatientID, schedule.ScheduleID, schedule.DoctorFullName)
	if err != nil {
		log.Println("Error inserting appointment:", err)
		http.Error(w, "Error creating appointment", http.StatusInternalServerError)
		return
	}

	// Retrieve the last inserted ID
	lastInsertedID, err := result.LastInsertId()
	if err != nil {
		log.Println("Error getting last inserted ID:", err)
		http.Error(w, "Error creating appointment", http.StatusInternalServerError)
		return
	}

	/*// Publish event to RabbitMQ
	err = PublishEvent(doctorID, patientID, "AppointmentCreated")
	if err != nil {
		log.Println("Failed to publish event:", err)
		// handle error
	}*/

	// Prepare and send the response
	response := model.ResponseWithID{
		Status:        http.StatusOK,
		Message:       "Appointment created successfully",
		AppointmentID: lastInsertedID,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

/*func PublishEvent(doctorID, patientID int, operation string) error {
	conn, err := config.ConnectRabbitMQ()
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// Declare the exchange
	err = ch.ExchangeDeclare(clinicReservationExchange, "topic", true, false, false, false, nil)
	if err != nil {
		return err
	}

	// Create the event structure
	event := map[string]interface{}{
		"doctorId":  doctorID,
		"patientId": patientID,
		"Operation": operation,
	}

	// Convert event to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Publish the event to the exchange
	err = ch.Publish(clinicReservationExchange, "reservation.event", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        eventJSON,
	})
	if err != nil {
		return err
	}

	return nil
}*/

func UpdateAppointment(w http.ResponseWriter, r *http.Request) {
	var request model.UpdateAppointmentRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db := config.Connect()
	defer db.Close()
	var existingAppointmentCount int
	err := db.QueryRow("SELECT COUNT(*) FROM appointments WHERE appointment_id = ? AND patient_id = ?", request.AppointmentID, request.PatientID).Scan(&existingAppointmentCount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if request.DoctorName != "" {
		var doctorID int
		err := db.QueryRow("SELECT doctor_id FROM doctors WHERE full_name = ?", request.DoctorName).Scan(&doctorID)
		if err != nil {
			http.Error(w, "Doctor not found", http.StatusNotFound)
			return
		}

		_, err = db.Exec("UPDATE appointments SET doctor_id = ? WHERE appointment_id = ?", doctorID, request.AppointmentID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if request.Date != "" && request.Hour != "" {
		_, err = db.Exec("UPDATE appointments SET appointment_date = ?, appointment_hour = ? WHERE appointment_id = ?", request.Date, request.Hour, request.AppointmentID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status": %d, "message": "Appointment updated successfully"}`, http.StatusOK)
}
func DeleteAppointment(w http.ResponseWriter, r *http.Request) {
	var request struct {
		AppointmentID int `json:"appointment_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db := config.Connect()
	defer db.Close()
	var patientID int
	err := db.QueryRow("SELECT patient_id FROM appointments WHERE appointment_id = ?", request.AppointmentID).Scan(&patientID)
	if err != nil {
		http.Error(w, "Appointment not found", http.StatusNotFound)
		return
	}
	_, err = db.Exec("DELETE FROM appointments WHERE appointment_id = ?", request.AppointmentID)
	if err != nil {
		http.Error(w, "Error deleting appointment", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status": %d, "message": "Appointment deleted successfully"}`, http.StatusOK)
}
func GetPatientAppointments(w http.ResponseWriter, r *http.Request) {
	// Extract the patient ID from the URL parameters
	patientID := mux.Vars(r)["patient_id"]

	// Connect to the database
	db := config.Connect()
	defer db.Close()

	// Query the database for all appointments associated with the patient
	query := "SELECT doctor_name, appointment_date, appointment_hour FROM appointments WHERE patient_id = ?"
	rows, err := db.Query(query, patientID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Define a struct to hold appointment information
	var appointments []model.Appointment
	for rows.Next() {
		var appointment model.Appointment
		err := rows.Scan(&appointment.DoctorName, &appointment.Date, &appointment.Hour)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		appointments = append(appointments, appointment)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the response content type and status
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode the appointments as JSON and send them as the response
	json.NewEncoder(w).Encode(appointments)
}
