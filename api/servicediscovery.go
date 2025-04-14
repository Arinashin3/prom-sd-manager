package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

type PrometheusLabels struct {
	Env            string `json:"__sd_env"`
	Local          string `json:"__sd_local"`
	HostGroup      string `json:"__sd_host_group"`
	Hostname       string `json:"__sd_hostname"`
	Ipaddr         string `json:"__sd_ipaddr"`
	OsType         string `json:"__sd_os_type"`
	ScrapePort     int64  `json:"__scrape_port__"`
	ScrapeInterval int64  `json:"__scrape_interval__"`
	ScrapeTimeout  int64  `json:"__scrape_timeout__"`
	ScrapeSsl      int64  `json:"__scrape_ssl__"`
}

type DefaultOption struct {
	JobId           int64
	JobName         string
	DefaultPort     int64
	DefaultInterval int64
	DafaultTimeout  int64
	DafaultSsl      int64
}

type PrometheusTargets struct {
	Targets []string         `json:"targets"`
	Labels  PrometheusLabels `json:"labels"`
}

func GetTargets(jobName string, dbFile string) ([]byte, error) {
	db, _ := sql.Open("sqlite3", dbFile)
	db.Exec("PRAGMA foreign_keys = ON;")
	defer db.Close()

	var columns string
	var query string
	var t []PrometheusTargets

	// Serch Default Option
	var opt DefaultOption
	columns = "jobid, default_port, default_interval, default_timeout"
	query = "SELECT " + columns + " FROM joblist WHERE jobname == '" + jobName + "';"
	row := db.QueryRow(query)
	row.Scan(&opt.JobId, &opt.DefaultPort, &opt.DefaultInterval, &opt.DafaultTimeout)
	if opt.JobId == 0 {
		log.Println("cant find job: " + jobName)
		return nil, nil
	}

	// Serch Inventory
	columns = "env, local, host_group, hostname, ipaddr, os_type, scrape_port, scrape_interval, scrape_timeout"
	query = "SELECT " + columns + " FROM option LEFT JOIN inventory ON option.hostid = inventory.hostid WHERE jobid == " + strconv.Itoa(int(opt.JobId)) + ";"
	rows, err := db.Query(query)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	for rows.Next() {
		var l PrometheusLabels
		var target string
		err = rows.Scan(&l.Env, &l.Local, &l.HostGroup, &l.Hostname, &l.Ipaddr, &l.OsType, &l.ScrapePort, &l.ScrapeInterval, &l.ScrapeTimeout)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		if l.ScrapePort == -1 {
			l.ScrapePort = opt.DefaultPort
		}
		if l.ScrapeSsl == -1 {
			l.ScrapeSsl = opt.DafaultSsl
		}
		if l.ScrapeInterval == -1 {
			l.ScrapeInterval = opt.DefaultInterval

		}
		if l.ScrapeTimeout == -1 {
			l.ScrapeTimeout = opt.DafaultTimeout
		}
		if l.ScrapePort == 0 {
			target = l.Ipaddr
		} else {
			target = l.Ipaddr + ":" + strconv.Itoa(int(l.ScrapePort))
		}

		t = append(t, PrometheusTargets{Targets: []string{target}, Labels: l})
	}

	return json.Marshal(t)

}

func HandlerServiceDiscovery(w http.ResponseWriter, r *http.Request, dbFile *string) {
	jobName := r.Header.Get("JobName")
	if jobName == "" {
		return
	}

	pTargets, err := GetTargets(jobName, *dbFile)
	if err != nil {
		log.Println(err)
	}
	w.Header().Add("Content-Type", "application/json")
	_, _ = w.Write(pTargets)

	return
}
