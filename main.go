package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"prom-sd-manager/api"
)

// func CreateDB() {
//
//		db, err := sql.Open("sqlite3", "test.db")
//		query := "create table inventory(hostgroup varchar(30), hostname varchar(30), ipaddress varchar(20), os_type varchar(10));"
//	}
//
//	func HTTPHandler() http.Handler {
//		var handle http.Handler
//		return handle
//	}

type tableList struct {
	Name string
}

func checkDBFile(dbfile *string) {
	db, _ := sql.Open("sqlite3", *dbfile)
	defer db.Close()

	db.Exec("PRAGMA foreign_keys = ON;")

	tables := []string{"inventory", "option", "joblist"}
	var tb tableList

	for _, tName := range tables {
		row := db.QueryRow("SELECT name FROM sqlite_master WHERE (type='table' and name='" + tName + "');")
		row.Scan(&tb.Name)

		if tb.Name == tName {
			continue
		}

		log.Println("create table: " + tName)
		switch tName {
		case "inventory":
			db.Exec("CREATE TABLE inventory(hostid integer primary key autoincrement, env varchar(10) default prd, local varchar(10) default kr,host_group varchar(30) default 'null', hostname varchar(30) unique, ipaddr varchar(20) unique, os_type varchar(10));")
		case "option":
			db.Exec("CREATE TABLE option(hostid integer, jobid integer, scrape_port integer default -1, scrape_interval integer default -1, scrape_timeout integer default -1, scrape_ssl integer default -1, foreign key(hostid) references inventory(hostid) on delete cascade, foreign key (jobid) references joblist(jobid), primary key (hostid, jobid));")
		case "joblist":
			db.Exec("CREATE TABLE joblist(jobid integer primary key autoincrement, jobname varchar(10), default_port integer default 0, default_interval integer default 15, default_timeout integer default 15, default_ssl integer default 0);")
			db.Exec("INSERT INTO joblist(jobname, default_port) VALUES('node_exporter', 9100);")
			db.Exec("INSERT INTO joblist(jobname, default_port) VALUES('process_exporter', 9256);")
			db.Exec("INSERT INTO joblist(jobname, default_port) VALUES('windows_exporter', 9182);")
			db.Exec("INSERT INTO joblist(jobname, default_port, default_ssl) VALUES('spectrum_exporter', 7443, 1);")
		}
	}
	return
}

func main() {
	dbFile := "test.db"
	checkDBFile(&dbFile)
	http.HandleFunc("/api/register", func(w http.ResponseWriter, r *http.Request) {
		api.HandlerRegister(w, r, &dbFile)
	})
	http.HandleFunc("/api/service_discovery", func(w http.ResponseWriter, r *http.Request) {
		api.HandlerServiceDiscovery(w, r, &dbFile)
	})
	http.ListenAndServe(":9095", nil)

}
