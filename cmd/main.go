package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"recordcpu/readcpu"
	"recordcpu/scmd"
	"strings"
	// "time"
)

type timestampCputime struct {
	timestamp int64
	cputime   []string
}

var (
	psgrepcmdStr string
	// psgrepcmdStr          string = "ps -ef | grep kworker | grep -v grep"
	cpCmdstr        string
	zpoolSyncCmdstr string
	recordinterval  int
)

type lockmap struct {
	lock                sync.Mutex
	pidTimestampCPUTime map[string][]timestampCputime
}

func main() {
	if len(os.Args) != 5 {
		fmt.Println("program cpstr zpoolsyncstr psgrepcmdStr recordinterval(ms)")
		return
	} else {
		cpCmdstr = os.Args[1]
		zpoolSyncCmdstr = os.Args[2]
		psgrepcmdStr = os.Args[3]
		var err error
		recordinterval, err = strconv.Atoi(os.Args[4])
		fmt.Printf("%v\n%v\n%v\n%v\n", cpCmdstr, zpoolSyncCmdstr, psgrepcmdStr, os.Args[4])
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	var mylockmap lockmap
	mylockmap.pidTimestampCPUTime = make(map[string][]timestampCputime, 20)
	ch := make(chan bool, 1)
	start := time.Now()
	go func(chan bool) {
		for {
			select {
			case <-ch:
				time.Sleep(30 * time.Second)
				ch <- true
				return
			default:
			}
			retpidstrn, err := scmd.RunCmdwithPipe(psgrepcmdStr)
			if err != nil {
				log.Println(err)
				continue
			}
			retpidstr := strings.TrimSpace(retpidstrn)
			retpidlist := strings.Split(retpidstr, "\n")
			timestamp := time.Since(start).Microseconds()
			for _, v := range retpidlist {
				var i int
				for {
					if v[i] == ' ' {
						break
					}
					i++
				}
				for {
					if v[i] != ' ' {
						// first pid number
						break
					}
					i++
				}
				j := i
				for {
					if v[j] == ' ' {
						// end +1
						break
					}
					j++
				}
				mylockmap.lock.Lock()
				if _, ok := mylockmap.pidTimestampCPUTime[v[i:j]]; !ok {
					mylockmap.pidTimestampCPUTime[v[i:j]] = make([]timestampCputime, 0)
				}
				mylockmap.lock.Unlock()
				go func(pid string) {
					pcpu, err := readcpu.ProcessCpu(pid)
					if err != nil {
						fmt.Println(err)
						return
					}
					mylockmap.lock.Lock()
					defer mylockmap.lock.Unlock()
					tmp := mylockmap.pidTimestampCPUTime[pid]
					tmp = append(tmp, timestampCputime{timestamp, pcpu})
					mylockmap.pidTimestampCPUTime[pid] = tmp
				}(v[i:j])
			}
			time.Sleep(time.Duration(recordinterval) * time.Millisecond)
		}
	}(ch)

	// cp and sync
	cpstarttime := time.Now()
	cpres, err := scmd.RunCmdwithPipe(cpCmdstr)
	if err != nil {
		log.Println(err)
		log.Println(cpres)
		os.Exit(1)
	}
	syncres, err := scmd.RunCmdwithPipe(zpoolSyncCmdstr)
	if err != nil {
		log.Println(err)
		log.Println(syncres)
		os.Exit(1)
	}
	aftersync := time.Since(cpstarttime)
	fmt.Printf("cp -> sync time : %v\n", aftersync.Milliseconds())

	ch <- true
	<-ch
	mylockmap.lock.Lock()
	defer mylockmap.lock.Unlock()
	// print
	fmt.Println("------all result")
	for pid, ltc := range mylockmap.pidTimestampCPUTime {
		fmt.Println(pid)
		if len(ltc) <= 0 {
			continue
		}
		i := 0
		for _, tc := range ltc {
			fmt.Println(tc)
			tmp, err := readcpu.SubListCpuTime(ltc[i].cputime, tc.cputime)
			if err != nil {
				fmt.Println(err)
				continue
			}
			for _, v := range tmp {
				if v < 0 {
					fmt.Println("order error")
					break
				}
			}
		}
	}

	var lencputime int
	for _, ltc := range mylockmap.pidTimestampCPUTime {
		if len(ltc) > 0 {
			lencputime = len(ltc[0].cputime)
			break
		}
	}
	sumcputime := make([]int, lencputime)

	// cost time
	fmt.Println("cost time")
	for pid, ltc := range mylockmap.pidTimestampCPUTime {
		fmt.Println(pid)
		if len(ltc) <= 0 {
			fmt.Println(pid, "len <= 0")
			continue
		}
		if len(ltc) == 1 {
			fmt.Println(ltc[0].cputime)
			for i, v := range ltc[0].cputime {
				vv, _ := strconv.Atoi(v)
				sumcputime[i] += vv
			}
		} else {
			tmp, err := readcpu.SubListCpuTime(ltc[0].cputime, ltc[len(ltc)-1].cputime)
			if err != nil {
				fmt.Println(err)
				continue
			} else {
				fmt.Println(tmp)
				for i, v := range tmp {
					sumcputime[i] += v
				}
			}
		}

	}

	// sum
	fmt.Println("sum : ", sumcputime)
}
