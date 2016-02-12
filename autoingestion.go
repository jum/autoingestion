/*
 * This is an unpublished work copyright 2015 Jens-Uwe Mager
 * 30177 Hannover, Germany, jum@anubis.han.de
 */

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"time"
)

type Config struct {
	UserName string
	PassWord string
	Vendor   string
	Report   string
	DateType string
	SubType  string
}

const (
	ConfigFile = ".autoingestion"
)

const URLBASE = "https://reportingitc.apple.com/autoingestion.tft?"

var (
	conf Config
)

func main() {
	// preset some sane defaults
	conf.Report = "Sales"
	conf.DateType = "Monthly"
	conf.SubType = "Summary"

	user, err := user.Current()
	if err != nil {
		panic(err.Error())
	}
	configPath := filepath.Join(user.HomeDir, ConfigFile)
	//fmt.Printf("configPath = %v\n", configPath)
	f, err := os.Open(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v: %v\n", configPath, err)
	} else {
		dec := json.NewDecoder(f)
		err = dec.Decode(&conf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v: %v\n", configPath, err)
		}
		f.Close()
	}
	flag.StringVar(&conf.UserName, "user", conf.UserName, "iTunes account")
	flag.StringVar(&conf.PassWord, "pass", conf.PassWord, "password")
	flag.StringVar(&conf.Vendor, "vendor", conf.Vendor, "vendor ID")
	flag.StringVar(&conf.Report, "report", conf.Report, "report type: Sales or Newsstand")
	flag.StringVar(&conf.DateType, "datetype", conf.DateType, "date type: Daily, Weekly, Monthly or Yearly")
	flag.StringVar(&conf.SubType, "subtype", conf.SubType, "report subtype: Summary, Detailed or Opt-In")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] [date]\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(2)
	}
	flag.Parse()
	//fmt.Printf("conf %+v\n", conf)
	f, err = os.OpenFile(configPath, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0600)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v: %v\n", configPath, err)
	} else {
		enc := json.NewEncoder(f)
		err := enc.Encode(&conf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v: %v\n", configPath, err)
		}
		f.Close()
	}
	if flag.NArg() > 1 {
		flag.Usage()
	}
	var date string
	if flag.NArg() == 1 {
		date = flag.Arg(0)
	} else {
		date = time.Now().Add(-24 * time.Hour).Format("20060102")
	}
	//fmt.Printf("date = %s\n", date)
	r, err := http.PostForm(URLBASE, url.Values{"USERNAME": {conf.UserName},
		"PASSWORD":     {conf.PassWord},
		"VNDNUMBER":    {conf.Vendor},
		"TYPEOFREPORT": {conf.Report},
		"DATETYPE":     {conf.DateType},
		"REPORTTYPE":   {conf.SubType},
		"REPORTDATE":   {date},
	})
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()
	//fmt.Printf("r : %#v\n", r)
	if r.StatusCode/100 != 2 {
		panic(r.Status)
	}
	errMsg := r.Header.Get("Errormsg")
	if len(errMsg) > 0 {
		fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], errMsg)
		os.Exit(1)
	}
	fname := r.Header.Get("Filename")
	fmt.Printf("%s\n", fname)
	f, err = os.Create(fname)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = io.Copy(f, r.Body)
	if err != nil {
		panic(err)
	}
}
