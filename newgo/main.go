package main

import (
	"fmt"
	"log"
	"net/http"
	"new/new-go/controller"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	router := mux.NewRouter()

	// Allow only requests from http://localhost:3000
	allowedOrigins := []string{"*"}

	// Create a new CORS handler with custom configuration
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"}, // Add any additional headers your application uses
		AllowCredentials: true,
		Debug:            true, // Enable debug mode for more verbose logging
	})

	// Use the CORS handler with your router
	router.Use(corsHandler.Handler)

	// Handle OPTIONS requests (preflight requests)
	router.Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Define your routes
	router.HandleFunc("/signin", controller.Signin).Methods("POST")
	router.HandleFunc("/signup", controller.Signup).Methods("POST")
	router.HandleFunc("/set-schedule", controller.SetSchedule).Methods("POST")
	router.HandleFunc("/doctors", controller.GetAvailableDoctors).Methods("GET")
	router.HandleFunc("/doctors/{full_name}/slots", controller.GetAvailableSlotsForDoctor).Methods("GET")
	router.HandleFunc("/book-slot", controller.ChooseSlot).Methods("POST")
	router.HandleFunc("/make-appointment", controller.CreateAppointment).Methods("POST")
	router.HandleFunc("/update-appointment", controller.UpdateAppointment).Methods("PUT")
	router.HandleFunc("/delete-appointment", controller.DeleteAppointment).Methods("DELETE")
	router.HandleFunc("/appointments/{patient_id}", controller.GetPatientAppointments).Methods("GET")

	fmt.Println("Connected to port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
