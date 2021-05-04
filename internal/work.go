package tablessh

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/fatih/color"
)

type Work struct {
	host     string
	commands []string
}

func CreateWorkload(table [][]string) chan Work {
	wl := []Work{}
	hosts := map[string]struct{}{}

	for _, row := range table {
		if _, ok := hosts[row[0]]; ok {
			log.Fatalf("duplicate host found: %v", row[0])
		}
		hosts[row[0]] = struct{}{}
		wl = append(wl, Work{host: row[0], commands: row[1:]})
	}

	if len(wl) > 1000 {
		log.Fatalf("that's a lot of work, maybe split them up")
	}

	c := make(chan Work, len(wl))
	for _, w := range wl {
		c <- w
	}
	close(c)
	return c
}

func DoWork(q chan Work, done chan struct{}, logDir string) {
	defer close(done)

getMoreWork:
	for w := range q {
		for i := 0; i < len(w.commands); i++ {
			cmd := exec.Command("ssh", "-o StrictHostKeyChecking=no", w.host, w.commands[i])
			log.Println(color.CyanString(fmt.Sprintf("%s step %d started", w.host, i)))
			out, err := cmd.CombinedOutput()
			writeLog(logDir+"/"+w.host, out)
			if err != nil {
				log.Println(color.RedString(fmt.Sprintf("%s run failed on step %d: %s", w.host, i, err.Error())))
				os.Rename(logDir+"/"+w.host, logDir+"/fail/"+w.host)
				continue getMoreWork
			}
		}
		os.Rename(logDir+"/"+w.host, logDir+"/success/"+w.host)
		log.Println(color.GreenString(fmt.Sprintf("%s run finished", w.host)))
	}
}

func writeLog(logfile string, data []byte) {
	f, err := os.OpenFile(logfile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Println(color.RedString(fmt.Sprintf("can't open log: %s", err.Error())))
		return
	}
	_, err = f.Write(data)
	if err != nil {
		log.Println(color.RedString(fmt.Sprintf("can't write log: %s", err.Error())))
		return
	}
	err = f.Close()
	if err != nil {
		log.Println(color.RedString(fmt.Sprintf("can't close log: %s", err.Error())))
		return
	}
}
