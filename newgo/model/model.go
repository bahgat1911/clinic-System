package model

type UserType string

const (
	Doctor  UserType = "Doctor"
	Patient UserType = "Patient"
)

type DoctorS struct {
	DoctorID int    `json:"doctor_id"`
	FullName string `json:"full_name"`
}

type User struct {
	UserID    int      `json:"user_id"`
	DoctorID  int      `json:"doctor_id"`
	PatientID int      `json:"patient_id"`
	Username  string   `json:"username"`
	Password  string   `json:"password"`
	Email     string   `json:"email"`
	UserType  UserType `json:"user_type"`
}

type Response struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    []User `json:"data"`
	Token   string `json:"token"`
}

type ScheduleRequest struct {
	ScheduleID     int    `json:"schedule_id"`
	DoctorID       int    `json:"doctor_id"`
	Date           string `json:"date"`
	Hour           string `json:"hour"`
	Status         string `json:"status"`
	PatientID      int    `json:"patient_id"`
	DoctorFullName string `json:"full_name"`
}

type CreateAppointmentRequest struct {
	DoctorName string `json:"doctor_name"`
	Date       string `json:"date"`
	Hour       string `json:"hour"`
	PatientID  int    `json:"patient_id"`
	Status     string `json:"status"`
	ScheduleID int    `json:"schedule_id"`
}

type UpdateAppointmentRequest struct {
	AppointmentID int    `json:"appointment_id"`
	DoctorName    string `json:"doctor_name"`
	Date          string `json:"date"`
	Hour          string `json:"hour"`
	PatientID     int    `json:"patient_id"`
}

type Appointment struct {
	AppointmentID int    `json:"appointment_id"`
	DoctorName    string `json:"doctor_name"`
	Date          string `json:"date"`
	Hour          string `json:"hour"`
	PatientID     int    `json:"patient_id"`
	Status        string `json:"status"`
	ScheduleID    int    `json:"schedule_id"`
	DoctorID      int    `json:"doctor_id"`
}

// model/response.go

// ResponseWithID is a generic response structure with an AppointmentID field
type ResponseWithID struct {
	Status        int    `json:"status"`
	Message       string `json:"message"`
	AppointmentID int64  `json:"appointment_id"`
}
