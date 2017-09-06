package main

import (
	"fmt"
	"time"
)

//App struct describing autoscaling app
type App struct {
	AppID    string `json:"app_id"`
	MaxCPU   int    `json:"max_cpu"`
	MinMem   int    `json:"min_mem"`
	Method   string `json:"method"`
	Interval int    `json:"interval"`
}

//Apps - all monitored apps
type Apps []App

//StartMonitor starts a ticker goroutine
func (a *App) StartMonitor() {
	tickers[a.AppID] = time.NewTicker(time.Second * time.Duration(a.Interval))
	go a.doMonitor()
}

//doMonitor will be storing the intermediate state of the app metrics
func (a *App) doMonitor() {
	var internal int
	var cpu, mem float64
	for range tickers[a.AppID].C {
		if !client.AppExists(a) {
			fmt.Printf("%s not found in /service/marathon/v2/apps\n", a.AppID)
			continue
		}
		//fmt.Printf("*** ticker:%s ", t)
		internal++
		cpu, mem = a.getCPUMem()
		fmt.Printf("*** app:%s ", a.AppID)
		fmt.Printf("cpu:%f, mem:%f\n", cpu, mem)
	}
}

//StopMonitor stops the ticker associated with the given app
func (a *App) StopMonitor() {
	tickers[a.AppID].Stop()
}

func (a *App) getCPUMem() (float64, float64) {

	marathonApp := client.GetMarathonApp(a.AppID)
	//fmt.Println(marathonApp)

	var stats1, stats2 TaskStats
	var cpu, cpu1, cpu2, cpuD, timeD float64
	var mem float64
	for _, task := range marathonApp.App.Tasks {
		//fmt.Printf("id:%s app_id:%s slave_id:%s\n", task.ID, task.AppID, task.SlaveID)
		stats1 = client.GetTaskStats(task.ID, task.SlaveID)
		//fmt.Println(stats)
		time.Sleep(time.Second * 1)
		stats2 = client.GetTaskStats(task.ID, task.SlaveID)

		cpu1 = stats1.Statistics.CpusSystemTimeSecs + stats1.Statistics.CpusUserTimeSecs
		cpu2 = stats2.Statistics.CpusSystemTimeSecs + stats2.Statistics.CpusUserTimeSecs
		cpuD = cpu2 - cpu1
		timeD = stats2.Statistics.Timestamp - stats1.Statistics.Timestamp
		cpu = cpu + (cpuD / timeD)
		mem = mem + (stats1.Statistics.MemRssBytes / stats1.Statistics.MemLimitBytes)
		//fmt.Printf("cpu:%f, mem:%f\n", cpu, mem)
	}
	cpu = cpu / float64(len(marathonApp.App.Tasks)) * 100
	mem = mem / float64(len(marathonApp.App.Tasks)) * 100
	return cpu, mem
}
