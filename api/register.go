package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
)

type GroupLabels struct {
	Task      string `json:"task"`
	Env       string `json:"env"`
	Local     string `json:"local"`
	HostGroup string `json:"host_group"`
	Hostname  string `json:"hostname"`
	Ipaddr    string `json:"Ipaddr"`
	OsType    string `json:"os_type"`
	Jobname   string `json:"jobname"`
	Jobport   string `json:"jobport"`
}

type Option struct {
	Hostid      int64
	Jobid       int64
	DefaultPort int64
}

type Data struct {
	GroupLabels GroupLabels `json:"groupLabels"`
}

func HandlerRegister(w http.ResponseWriter, r *http.Request, dbFile *string) {
	//b, err := r.GetBody()

	d, err := os.ReadFile("testdata.json")
	var data Data
	var opt Option
	json.Unmarshal(d, &data)
	log.Println(data, err)
	l := data.GroupLabels

	db, _ := sql.Open("sqlite3", *dbFile)
	db.Exec("PRAGMA foreign_keys = ON;")
	defer db.Close()
	var query string
	query = "INSERT INTO inventory (env, local, host_group, hostname, ipaddr, os_type) VALUES ('" + l.Env + "','" + l.Local + "','" + l.HostGroup + "','" + l.Hostname + "','" + l.Ipaddr + "','" + l.OsType + "');"
	_, err = db.Exec(query)
	if err != nil {
		log.Println(err)
		return
	}
	row := db.QueryRow("SELECT hostid FROM inventory WHERE hostname == '" + l.Hostname + "';")
	err = row.Scan(&opt.Hostid)
	if err != nil {
		log.Println(err)
		return
	}
	row = db.QueryRow("SELECT jobid FROM joblist WHERE jobname == '" + l.Jobname + "';")
	err = row.Scan(&opt.Jobid)
	if err != nil {
		log.Println(err)
		return
	}

	_, err = db.Exec("INSERT INTO option (hostid, jobid) VALUES (" + strconv.Itoa(int(opt.Hostid)) + "," + strconv.Itoa(int(opt.Jobid)) + ");")
	if err != nil {
		log.Println(err)
		return
	}

	w.Write(nil)
	return

}
