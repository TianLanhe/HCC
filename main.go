package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Record struct {
	Name string
	Set  IntSet
}

type Result struct {
	Prefix string
	Num    string
	N      int
}

// 配置参数
var x, y, n, concurrentNum int
var shouldPrint bool // 是否打印过程

var filter map[string]int = make(map[string]int) // 重复结果
var results []Result                             // 缓存结果，已去重
var total uint64                                 // 结果总数
var mutex sync.Mutex
var ioTime int64     // io总时间
var outFile *os.File // 输出文件句柄

var maxFlushLength int // 结果达到一定量进行刷新

var resultChan chan IntSet // 结果通道
var exitChan chan struct{} // 退出信号

var totalSet IntSet

const (
	InputFileName  string = "data.txt"
	OutputFileName string = "output.txt"
	ConfigFileName string = "config.json"

	ChanLength = 10000 // 通道长度
	PrintSize = 500000 // 打印提示的间隔
)

func OutputAResulst(set IntSet) {
	resultChan <- set
}

func DealOneResult(set IntSet) {
	tempSet := totalSet.Copy()
	tempSet.DifferentWith(set)

	result := Result{
		Num: fmt.Sprintf("%v", tempSet),
		N:   x - set.Len(),
	}

	if _, ok := filter[result.Num]; !ok {
		filter[result.Num] = 1

		if len(results) == maxFlushLength {
			go flushToFile(results)
			results = make([]Result, 0, maxFlushLength+1)
		}
		results = append(results, result)
	} else {
		filter[result.Num]++
	}
}

func pushSet(preSet, curSet *IntSet, m []*Record, index int) {
	*preSet = *curSet
	curSet.UnionWith(m[index].Set)
}

func popSet(preSet, curSet *IntSet, m []*Record, index int) {
	temp := *preSet
	*preSet = *curSet
	curSet.DifferentWith(m[index].Set)
	curSet.UnionWith(temp)
}

func BeginTask(m []*Record, n, maxCount int) {
	finishChan := make(chan struct{},concurrentNum)

	for i := 1; i < concurrentNum; i++ {
		go beginTask(m, n, maxCount, i,finishChan)
	}
	beginTask(m, n, maxCount, 0,nil)

	for i:=1;i<concurrentNum;i++{
		<-finishChan
	}
	close(finishChan)
}

func beginTask(m []*Record, n, maxCount, idx int,finish chan struct{}) {
	var stack Stack
	var setstack SetStack

	mSize := len(m)
	needPop := false
	count := 0	// 用来打印的计数器

	step := 1	// 步长

	stack.Push(0)
	setstack.Push(m[0].Set)
	for !stack.Empty() {
		if shouldPrint && count == 0 {
			fmt.Println(stack.Elems(), "total: ", len(filter))
			count++
		}

		if !needPop && stack.Len() == n {
			OutputAResulst(setstack.Top())
			needPop = true
		}

		last := stack.Top()
		stackSize := stack.Len()
		if last == mSize-1 || mSize-1-last < n- stackSize {
			stack.Pop()
			setstack.Pop()
			needPop = true

			count++
			if count == PrintSize {
				count = 0
			}
			continue
		}

		if needPop {
			stack.Pop()
			setstack.Pop()
			stackSize--
			needPop = false
		}

		if stackSize == n - 1 {
			if last + step + idx < mSize {
				index := last + step + idx

				stack.Push(index)

				if setstack.Empty() {
					setstack.Push(*m[index].Set.Copy())

					needPop = m[index].Set.Len() > maxCount
				} else {
					a := setstack.Top()
					temp := a.Copy()
					temp.UnionWith(m[index].Set)
					setstack.Push(*temp)

					needPop = temp.Len() > maxCount
				}

				step = concurrentNum
				idx = 0
			} else {
				step -= mSize - last - idx - 1
				needPop = true
			}
		}else if last+1 < mSize {
			stack.Push(last + 1)

			if setstack.Empty() {
				setstack.Push(*m[last+1].Set.Copy())

				needPop = m[last+1].Set.Len() > maxCount
			} else {
				a := setstack.Top()
				temp := a.Copy()
				temp.UnionWith(m[last+1].Set)
				setstack.Push(*temp)

				needPop = temp.Len() > maxCount
			}
		}
	}

	if finish != nil {
		finish <- struct{}{}
	}
}

func flushToFile(results []Result) {
	mutex.Lock()
	startTime := time.Now()

	for i := 0; i < len(results); i++ {
		fmt.Fprintf(outFile, "%s\r\n", results[i].Num)
	}

	milis := (time.Now().UnixNano() - startTime.UnixNano()) / 100000

	fmt.Printf("本次输出 %d 条结果\n", len(results))
	fmt.Printf("本次I/O耗时 %d ms\n", milis)

	total += uint64(len(results))
	ioTime += milis

	fmt.Printf("累计输出 %d 条结果\n", total)
	fmt.Printf("累计I/O耗时 %d ms\n\n", ioTime)

	mutex.Unlock()
}

func startOuputRuntine() {
	go func() {
		for {
			select {
			case result := <-resultChan:
				DealOneResult(result)
			case <-exitChan:
				for {
					select {
					case result := <-resultChan:
						DealOneResult(result)
					default:
						flushToFile(results)
						exitChan <- struct{}{}
						return
					}
				}
			}
		}
	}()
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// 读取配置
	if bytes, err := ioutil.ReadFile(ConfigFileName); err != nil {
		fmt.Printf("read configuration file error:%v, filename:%v", err, ConfigFileName)
		os.Exit(1)
	} else {
		config := make(map[string]int)
		json.Unmarshal(bytes, &config)
		x = config["X"]
		y = config["Y"]
		n = config["N"]

		maxFlushLength = config["MaxFlushLength"]

		if config["Print"] != 0 {
			shouldPrint = true
		} else {
			shouldPrint = false
		}

		concurrentNum = config["Concurrent"]
		if concurrentNum == 0 {
			concurrentNum = runtime.NumCPU()
		}
	}

	// 初始化全局变量
	var err error
	results = make([]Result, 0, maxFlushLength+1)
	filter = make(map[string]int)
	total = 0
	ioTime = 0
	if outFile, err = os.Create(OutputFileName); err != nil {
		fmt.Printf("create output file error:%v, filename:%v", err, OutputFileName)
		os.Exit(1)
	}
	for i := 1; i <= x; i++ {
		totalSet.Add(i)
	}

	exitChan = make(chan struct{})
	resultChan = make(chan IntSet, ChanLength)

	// 打开输入文件
	file, err := os.Open(InputFileName)
	if err != nil {
		fmt.Printf("Open File Error:%v\n", err)
		os.Exit(1)
	}

	// 读取文件
	m := []*Record{}
	count := uint64(0)
	rd := bufio.NewReader(file)
	for {
		//以'\n'为结束符读入一行
		line, err := rd.ReadString('\n')
		line = strings.Trim(line, "\r\n")

		if (err != nil || io.EOF == err) && len(line) == 0 {
			break
		}

		count++
		slice := strings.Split(line, ",")
		var set IntSet
		for _, num := range slice {
			n, err := strconv.ParseInt(num, 10, 32)
			if err != nil {
				fmt.Printf("string convert error:%v, num:%v\n", err, num)
				continue
			} else if n > int64(x) {
				fmt.Printf("i:%d, n:%v, x:%v\n", count, n, x)
				continue
			}
			set.Add(int(n))
		}
		m = append(m, &Record{strconv.FormatUint(count, 10), set})
	}
	file.Close()

	// 读取完毕
	for _, record := range m {
		fmt.Printf("%s:%s\n", record.Name, record.Set)
	}
	fmt.Printf("文件读取完毕！共读取 %d 行\n", len(m))

	// 记录耗时
	// defer func() func() {
	// 	start := time.Now()
	// 	return func() {
	// 		fmt.Printf("耗时 %v s\n", time.Now().Sub(start).Seconds())
	// 	}
	// }()()

	// 开启输出协程专门负责输出文件
	startOuputRuntine()

	startTime := time.Now()

	// 开始处理
	BeginTask(m, n, x-y)

	endTime := time.Now()

	exitChan <- struct{}{}
	<-exitChan

	fmt.Printf("运行耗时：%d min %d sec\n", (endTime.Unix()-startTime.Unix())/60, (endTime.Unix()-startTime.Unix())%60)
	fmt.Printf("I/O耗时：%d ms\n", ioTime)
	fmt.Printf("输出结果数量：%d\n", total)
	repeat := uint64(0)
	for _, val := range filter {
		repeat += uint64(val)
	}
	fmt.Printf("筛除重复结果数量：%d\n", repeat-total)

	outFile.Close()
}
