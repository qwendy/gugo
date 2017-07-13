---
title: lec01学习笔记
date: 2017-06-28 11:24:01
tags:
    - MapReduce
    - 笔记
category:
    - 6.824课程
---
分为: 存储，通讯和计算
关键点：实现，性能，容错和一致性。一致性和性能不可兼得
split-brain的意思是A true split brain means multiple systems are online and have accessed an exclusive resource simultaneously

MapReduce： 定义Map和Reduce函数，输入数组形式的数据。

扩展：不互相等待，不共享数据。可并发执行。可单纯的配置更多的计算机来获得更多的吞吐量
性能：Hard to build a network than can run 1000x faster than a single computer.So they cared about minimizing movement of data over the network.一般会被网络限制。尽量减少数据在网络上传输。
容错：当Map和Reduce失败时，重新运行。他们仅仅是函数而已--他们不修改输入的内容，不用保存状态，不共享内存，没有Map-Map和Reduce-Reduce的交互。所以重新执行会得到相同的输出。纯函数（pure function）的这个需求是相对于其他并行编程方案的主要限制，也是MR简单的原因。

MapReduce由Mater和Worker组成。Master给workers分配工作，决定worker运行map方法还是reduce方法。当worker执行错误，直接重启worker就好。
master的调度程序

``` go
func (mr *Master) schedule(phase jobPhase) {
	var ntasks int
	var nios int // number of inputs (for reduce) or outputs (for map)
	switch phase {
	case mapPhase:
		ntasks = len(mr.files)
		nios = mr.nReduce
	case reducePhase:
		ntasks = mr.nReduce
		nios = len(mr.files)
	}
	fmt.Printf("Schedule: %v %v tasks (%d I/Os)\n", ntasks, phase, nios)
	var wg sync.WaitGroup
	for i := 0; i < ntasks; {
		wg.Add(1)
		go func(i, nios int, phase jobPhase) {
			defer wg.Done()
			workrName := <-mr.registerChannel
			for true {
				ok := call(workrName, "Worker.DoTask", &DoTaskArgs{
					JobName:       mr.jobName,
					File:          mr.files[i],
					Phase:         phase,
					TaskNumber:    i,
					NumOtherPhase: nios,
				}, new(struct{}))
				if ok {
					go func() {
						mr.registerChannel <- workrName
					}()
					return
				}
			}

		}(i, nios, phase)
	}
	wg.Wait()
	fmt.Printf("Schedule: %v phase done\n", phase)
}
```
map方法读入input数据，使用mapF将数据处理之后，将处理后的数据存在指定的中间文件中，中间文件的文件名决定此文件将由哪个reduce执行。

``` go
func doMap(
	jobName string, // the name of the MapReduce job
	mapTaskNumber int, // which map task this is
	inFile string,
	nReduce int, // the number of reduce task that will be run ("R" in the paper)
	mapF func(file string, contents string) []KeyValue,
) {
	content, err := ioutil.ReadFile(inFile)
	if err != nil {
		log.Fatal("read input file content ", err)
	}
	kvs := mapF(inFile, string(content))
	files := []*os.File{}
	for i := 0; i < nReduce; i++ {
		fileName := reduceName(jobName, mapTaskNumber, i)
		file, err := os.Create(fileName)
		if err != nil {
			log.Fatal("Create File", err)
		}
		defer file.Close()
		files = append(files, file)
	}
	for _, kv := range kvs {
		e := json.NewEncoder(files[int(ihash(kv.Key))%nReduce])
		e.Encode(&kv)
	}
}
```

reduce读取指定的中间文件，使用reduceF处理后，输出结果。

``` go
func doReduce(
	jobName string, // the name of the whole MapReduce job
	reduceTaskNumber int, // which reduce task this is
	nMap int, // the number of map tasks that were run ("M" in the paper)
	reduceF func(key string, values []string) string,
) {
	kvMap := make(map[string][]string)

	// 读取同一个 reduce task 下的所有文件，保存至哈希表
	for i := 0; i < nMap; i++ {
		filename := reduceName(jobName, i, reduceTaskNumber)
		f, err := os.Open(filename)
		if err != nil {
			log.Fatal("Open File Err: ", filename)
		}
		defer f.Close()
		decoder := json.NewDecoder(f)
		var kv KeyValue
		for decoder.More() {
			err := decoder.Decode(&kv)
			if err != nil {
				log.Fatal("Json decode failed, ", err)
			}
			kvMap[kv.Key] = append(kvMap[kv.Key], kv.Value)
		}
	}

	// 对哈希表所有 key 进行升序排序
	keys := []string{}
	for k, _ := range kvMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	mergeFile, err := os.Create(mergeName(jobName, reduceTaskNumber))
	if err != nil {
		log.Fatal("Create Merge File Err: ", err)
	}
	defer mergeFile.Close()
	encoder := json.NewEncoder(mergeFile)
	for _, k := range keys {
		encoder.Encode(KeyValue{k, reduceF(k, kvMap[k])})
	}
}
```

(课程),[https://github.com/qwendy/Distributed-Systems/tree/master/Lec01_Introduction] 
(实验的代码)[https://github.com/qwendy/Distributed-Systems/tree/master/6.824/src/mapreduce] 