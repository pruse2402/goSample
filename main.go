package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"gosample/server/conf"
	"gosample/server/dbcon"
	"gosample/server/routes"
)

func main() {
	/// Set flag along with std flags to print file name in log ///
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	/// Set log file ///
	//logFile, err := os.OpenFile("gosample.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0777)
	//if err != nil {
	//	fmt.Printf("Error in opening log file: %s", err.Error())
	//	os.Exit(1)
	//}
	//log.SetOutput(logFile)
	//
	//defer logFile.Close()

	/// DB Connection ///
	dbcon.ConnectMongoDB()
	defer dbcon.CloseDB()

	/// Router config ///
	router := routes.RouterConfig()

	/// Init scripts ///
	//initapp.Boot()

	server := http.Server{
	Addr:         fmt.Sprintf(":%d", conf.Port),
	ReadTimeout:  90 * time.Second,
	WriteTimeout: 90 * time.Second,
	Handler:      router,
	}

	log.Printf("Listening on: %d", conf.Port)

	err := server.ListenAndServe()
	if err != nil {
		log.Printf("Error in listening server: %s", err.Error())
	}
}