package readcpu

import (
	"bufio"
	"io"
	"log"
	"os"
	"recordcpu/scmd"
	"strconv"
	"strings"
)

func TotalCpu() ([]string, error) {
	result, err := scmd.RunCmdwithPipe("cat /proc/stat | grep \"cpu \"")
	// search "cpu ", not "cpu"
	if err != nil {
		log.Println(result)
		return nil, err
	}
	ret := strings.Split(result[5:], " ")
	return ret, nil
}

// l2 - l1
func SubListCpuTime(l1, l2 []string) ([]int, error) {
	ret := make([]int, 0)
	var v1, v2 int
	var err error
	for i, st := range l1 {
		v1, err = strconv.Atoi(st)
		if err != nil {
			return nil, err
		}
		v2, err = strconv.Atoi(l2[i])
		if err != nil {
			return nil, err
		}
		ret = append(ret, v2-v1)
	}
	return ret, nil
}

func ProcessCpu(pid string) ([]string, error) {
	statpath := "/proc/" + pid + "/stat"
	file, err := os.Open(statpath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buf := bufio.NewReader(file)
	// only first line
	line, err := buf.ReadString('\n')
	if err != nil && err != io.EOF {
		return nil, err
	}
	var result []string
	if len(line) > 0 {
		result = strings.Split(line, " ")
	}
	ret := make([]string, 4)
	for i := 0; i < 4; i++ {
		ret[i] = result[13+i]
	}
	return ret, nil
}
