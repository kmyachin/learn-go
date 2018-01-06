package main

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

const (
	quotaLimit = 1
)

func ExecutePipeline(job_list ...job) {
	ch_in := make(chan interface{}, 1)
	var ch_out chan interface{}
	for _, f := range job_list {
		ch_out = make(chan interface{}, 1)
		go func(f job, in, out chan interface{}) {
			f(in, out)
			close(out)
		}(f, ch_in, ch_out)
		ch_in = ch_out
	}
	<-ch_out
}

func DataCrc(data string, out chan<- string) {
	out <- DataSignerCrc32(data)
}

func SingleHashOne(w *sync.WaitGroup, data string, out chan interface{}, quota_ch chan struct{}) {
	defer w.Done()

	quota_ch <- struct{}{}
	md5 := DataSignerMd5(data)
	<-quota_ch

	//fmt.Println(data, " SingleHash md5(data) ", md5)
	ch_md5 := make(chan string)
	ch_data := make(chan string)

	go DataCrc(md5, ch_md5)
	go DataCrc(data, ch_data)

	crc := <-ch_data
	//fmt.Println(data, " SingleHash crc32(data) ", crc)
	crc_md5 := <-ch_md5
	//fmt.Println(data, " SingleHash crc32(md5(data)) ", crc_md5)
	result := crc + "~" + crc_md5
	//fmt.Println(data, " SingleHash result ", result)
	out <- result
}

func SingleHash(in, out chan interface{}) {
	w := &sync.WaitGroup{}
	quota_ch := make(chan struct{}, quotaLimit)
	for s := range in {
		data := fmt.Sprint(s)
		//fmt.Println(data, " SingleHash data", data)

		w.Add(1)
		go SingleHashOne(w, data, out, quota_ch)
	}
	w.Wait()
}

func MultiHashOne(w *sync.WaitGroup, data string, out chan interface{}) {
	defer w.Done()
	var result string
	var ch [6]chan string
	for i := 0; i < 6; i++ {
		ch[i] = make(chan string)
		go DataCrc(fmt.Sprint(i, data), ch[i])
	}

	for i := 0; i < 6; i++ {
		v := <-ch[i]
		//fmt.Println(data, " MultiHash crc32(th+step1)", i, v)
		result += v
	}

	//fmt.Println(data, " MultiHash result: ", result)
	out <- result
}

func MultiHash(in, out chan interface{}) {
	w := &sync.WaitGroup{}
	for s := range in {
		data := fmt.Sprint(s)
		w.Add(1)
		go MultiHashOne(w, data, out)
	}
	w.Wait()
}

func CombineResults(in, out chan interface{}) {
	var sl []string
	for s := range in {
		sl = append(sl, fmt.Sprint(s))
	}
	sort.Strings(sl)
	out <- strings.Join(sl, "_")
}

func main() {
	inputData := []int{0, 1}
	jobs := []job{
		job(func(in, out chan interface{}) {
			for _, fibNum := range inputData {
				out <- fibNum
			}
		}),
		job(SingleHash),
		job(MultiHash),
		job(CombineResults),
	}
	ExecutePipeline(jobs...)
	println("run as\n\ngo test -v -race")
}
