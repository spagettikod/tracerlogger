package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/spagettikod/gotracer"
	"github.com/spagettikod/sdb"
)

const (
	SDBDomainName   = "tracerlogger"
	DailyPVPowerURI = "/day/pv/power"
)

var (
	port, accessKey, secretKey string
	db                         sdb.SimpleDB
	ErrNoRowsFound             error = errors.New("No rows found")
)

func init() {
	flag.StringVar(&port, "p", "", "COM port where EPsolar Tracer is connected")
	flag.StringVar(&accessKey, "a", "", "AWS Access Key to be able to connect to SimpleDB")
	flag.StringVar(&secretKey, "s", "", "AWS Secret Key, to be able to connect to SimpleDB")
}

func setup() {
	db = sdb.NewSimpleDB(accessKey, secretKey, sdb.SDBRegionEUWest1)
	lsr, err := db.ListDomains()
	if err != nil {
		log.Fatalf("tracelogger.setup - list sdb domains failed: %v", err)
	}
	var exists = false
	for _, d := range lsr.DomainNames {
		exists = d == SDBDomainName
	}
	if !exists {
		_, err = db.CreateDomain(SDBDomainName)
		if err != nil {
			log.Fatalf("tracelogger.setup - create sdb domain failed: %v", err)
		}
	}
}

func doLog() error {
	t, err := gotracer.Status(port)
	if err != nil {
		log.Printf("tracerlogger.doStatus - reading tracer status failed: %v", err)
		return err
	}
	var b []byte
	b, err = json.Marshal(t)
	if err != nil {
		log.Printf("tracerlogger.doStatus - json marshal of read data failed: %v", err)
		return err
	}

	item := sdb.NewItem(strconv.FormatInt(t.Timestamp.Unix(), 10))
	item.AddAttribute("data", string(b))
	_, err = db.PutAttributes(SDBDomainName, item)
	if err != nil {
		log.Printf("tracerlogger.doStatus - putting item into simpledb failed: %v", err)
		return err
	}
	return err
}

func LatestLogHandler(w http.ResponseWriter, req *http.Request) {
	resp, err := db.Select("SELECT data FROM " + SDBDomainName + " WHERE ItemName() > '0' ORDER BY ItemName() DESC LIMIT 1")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(resp.Items) == 0 {
		http.Error(w, ErrNoRowsFound.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, resp.Items[0].Attributes[0].Value)
}

func uriDate(uriRoot string, req *http.Request) (t time.Time, err error) {
	id := req.URL.RequestURI()[len(uriRoot):]
	if id == "" {
		t = time.Now().UTC()
	} else {
		t, err = time.Parse("2006-01-02", id[1:]) // Remove the leading slash
	}
	return
}

func startOfDay(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func endOfDay(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, int(time.Second-1), t.Location())
}

func fetch(from, to time.Time, v interface{}) error {
	// Set date to begining and end of day to get the
	from = startOfDay(from)
	to = endOfDay(to)
	q := fmt.Sprintf("SELECT data FROM %v WHERE ItemName() BETWEEN '%v' AND '%v' ORDER BY ItemName() ASC", SDBDomainName, from.Unix(), to.Unix())

	//var resp sdb.SelectResponse
	resp, err := db.Select(q)
	if err != nil {
		return err
	}
	if len(resp.Items) == 0 {
		return ErrNoRowsFound
	}

	// Put all JSON data for each data point into an JSON array
	var timestamps []time.Time
	var data []byte
	data = append(data, []byte("[")...)
	for i, v := range resp.Items {
		data = append(data, []byte(v.Attributes[0].Value)...)
		// Skip ',' on the last item
		if i < len(resp.Items)-1 {
			data = append(data, []byte(",")...)
		}
		timestamp, err := strconv.ParseInt(v.Name, 10, 64)
		if err != nil {
			return err
		}
		timestamps = append(timestamps, time.Unix(timestamp, 0))
	}
	data = append(data, []byte("]")...)

	// Unmarshal JSON data point array into an TracerStatus array
	err = json.Unmarshal(data, &v)
	if err != nil {
		return err
	}

	return err
}

type TracerPowerStatus struct {
	ArrayPower float32   `json:"pvp"`
	Timestamp  time.Time `json:"t"`
}

func DayPowerHandler(w http.ResponseWriter, req *http.Request) {
	t, err := uriDate(DailyPVPowerURI, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	var v []TracerPowerStatus
	var b []byte
	err = fetch(t, t, &v)
	b, err = json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	fmt.Fprint(w, string(b))
}

func main() {
	flag.Parse()
	if port == "" || accessKey == "" || secretKey == "" {
		flag.PrintDefaults()
		os.Exit(-1)
	}
	setup()
	go func() {
		c := time.Tick(5 * time.Second)
		for _ = range c {
			doLog()
		}
	}()
	/*http.HandleFunc("/", LatestLogHandler)
	http.HandleFunc(DailyPVPowerURI, DayPowerHandler)
	http.HandleFunc(DailyPVPowerURI+"/", DayPowerHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("ListenAndServe: %v", err)
	}*/
}
