package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fatih/color"
	tablessh "github.com/nalcire/tablessh/internal"
)

var workers int

func init() {
	flag.IntVar(&workers, "w", 100, "number of workers")
	flag.Parse()

	if workers == 0 || workers > 300 {
		log.Fatal("don't be crazy now, use reasonable workers")
	}
}

func main() {
	if len(flag.Args()) != 1 {
		fmt.Println("usage: tablessh [csv file]")
		fmt.Printf("%+v\n", os.Args)
		os.Exit(1)
	}
	f, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err.Error())
	}
	r := csv.NewReader(f)
	table, err := r.ReadAll()
	if err != nil {
		log.Fatal(err.Error())
	}

	q := tablessh.CreateWorkload(table)
	logDir := flag.Arg(0) + "_results"
	err = os.Mkdir(logDir, 0755)
	if err != nil {
		log.Fatal(err.Error())
	}
	err = os.Mkdir(logDir+"/success", 0755)
	if err != nil {
		log.Fatal(err.Error())
	}
	err = os.Mkdir(logDir+"/fail", 0755)
	if err != nil {
		log.Fatal(err.Error())
	}

	waits := []chan struct{}{}
	for i := 0; i < workers; i++ {
		done := make(chan struct{})
		waits = append(waits, done)
		go tablessh.DoWork(q, done, logDir)
	}

	go func() {
		time.Sleep(1 * time.Minute)
		log.Println(color.YellowString("there are [%d] items left", len(q)))
	}()

	for _, w := range waits {
		<-w
	}
}
