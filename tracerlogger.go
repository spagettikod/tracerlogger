package main

import (
	"database/sql"
	"flag"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spagettikod/gotracer"
)

const (
	CreateLogTable string = `
		CREATE TABLE IF NOT EXISTS log (
			timestamp			NUMERIC PRIMARY KEY,
			array_voltage 		REAL,
			array_current 		REAL,
			array_power 		REAL,
			battery_voltage 	REAL,
			battery_current 	REAL,
			battery_soc 		INTEGER,
			battery_temp 		REAL,
			battery_max_volt 	REAL,
			battery_min_volt 	REAL,
			device_temp 		REAL,
			load_voltage 		REAL,
			load_current 		REAL,
			load_power 			REAL,
			load 				NUMERIC,
			consumed_day 		REAL,
			consumed_month 		REAL,
			consumed_year		REAL,
			consumed_total 		REAL,
			generated_day 		REAL,
			generated_month 	REAL,
			generated_year		REAL,
			generated_total 	REAL
		);
	`
	InsertStmt string = `
		INSERT INTO log (
			timestamp,
			array_voltage,
			array_current,
			array_power,
			battery_voltage,
			battery_current,
			battery_soc,
			battery_temp,
			battery_max_volt,
			battery_min_volt,
			device_temp,
			load_voltage,
			load_current,
			load_power,
			load,
			consumed_day,
			consumed_month,
			consumed_year,
			consumed_total,
			generated_day,
			generated_month,
			generated_year,
			generated_total
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?);
	`
)

var (
	port, dbFile string
	db           *sql.DB
)

func init() {
	flag.StringVar(&port, "p", "", "COM port where EPsolar Tracer is connected")
	flag.StringVar(&dbFile, "db", "", "path and filename of SQLite database")
}

func setup() (err error) {
	db, err = sql.Open("sqlite3", dbFile)
	if err != nil {
		return
	}

	_, err = db.Exec(CreateLogTable)
	return
}

func doLog() error {
	t, err := gotracer.Status(port)
	if err != nil {
		log.Printf("reading tracer status failed: %v", err)
		return err
	}
	_, err = db.Exec(InsertStmt, t.Timestamp, t.ArrayVoltage, t.ArrayCurrent, t.ArrayPower, t.BatteryVoltage,
		t.BatteryCurrent, t.BatterySOC, t.BatteryTemp, t.BatteryMaxVoltage, t.BatteryMinVoltage, t.DeviceTemp,
		t.LoadVoltage, t.LoadCurrent, t.LoadPower, t.Load, t.EnergyConsumedDaily, t.EnergyConsumedMonthly,
		t.EnergyConsumedAnnual, t.EnergyConsumedTotal, t.EnergyGeneratedDaily, t.EnergyGeneratedMonthly,
		t.EnergyGeneratedAnnual, t.EnergyGeneratedTotal)

	return err
}

func main() {
	flag.Parse()
	if port == "" || dbFile == "" {
		flag.PrintDefaults()
		os.Exit(-1)
	}
	err := setup()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	doLog()
	c := time.Tick(5 * time.Second)
	for _ = range c {
		doLog()
	}
}
