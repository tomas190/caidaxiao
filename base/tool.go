package base

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"time"

	"gopkg.in/mgo.v2/bson"
)

const (
	Debug_MODE     = true //是否印 DeBug LOG
	DateLayout     = "2006-01-02"
	TimeLayout     = "15:04:05"
	DateTimeLayout = "2006-01-02 15:04:05"
)

// 取亂數 1~num
func RanInt(num int) int {
	if num < 0 {
		fmt.Printf("传入了负数 %d", num)
		return 0
	}

	if num == 0 {
		num++
	}
	rndInt := rand.Intn(num) + 1
	return rndInt
}

// TimeNowStr 輸出格式為 2019/11/4 20:15:26
func TimeNowStr() string {
	n := time.Now()
	timeStr := fmt.Sprintf("%d/%d/%d %d:%d:%d", n.Year(), n.Month(), n.Day(), n.Hour(), n.Minute(), n.Second())
	return timeStr
}

//  转化成"2006-01-02 15:04:05"的时间模版
// 輸入為time.Now().Unix()
func TimeFormatDate(timestamp int64) string {
	//timestamp := time.Now().Unix()
	datetime := time.Unix(timestamp, 0).Format(DateTimeLayout)
	return datetime
}

// 現在時間 格式為""2006-01-02""
func DateFromNow() string {
	return time.Now().Format(DateLayout)
}

// 透過time.Now().Unix()出來的秒數轉為 "2006-01-02"
func DateFromTimeStamp(timestamp int64) string {
	return time.Unix(timestamp, 0).Format(DateLayout)
}

// Int轉str 将整形转换成字符串
func IntToStr(n int) string {
	return strconv.FormatInt(int64(n), 10)
}

// Int轉str 将整形转换成字符串
func Int32ToStr(n int32) string {
	return strconv.FormatInt(int64(n), 10)
}

// Float轉Str 浮点数转换成字符串
func FloatToStr(f float64) string {
	return fmt.Sprintf("%f", f)
}

// 四捨五入取小數兩位
func RoundingTwo(value float64) float64 {
	value64, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", value), 64)
	return value64
}

// 四捨五入取小數四位
func RoundingFour(value float64) float64 {
	value64, _ := strconv.ParseFloat(fmt.Sprintf("%.4f", value), 64)
	return value64
}

// 四捨五入取小數七位(後端統一)
func RoundingSeven(value float64) float64 {
	value64, _ := strconv.ParseFloat(fmt.Sprintf("%.7f", value), 64)
	return value64
}

// DeBug模式下才會印出的log
func Debug_log(format string, a ...interface{}) {

	if Debug_MODE {
		msg := fmt.Sprintf(format, a...)
		_, file, line, ok := runtime.Caller(1)
		file = filepath.Base(file)
		if !ok {
			return
		}
		fmt.Print(TimeNowStr()+" "+file+":", line, ": "+msg+"\n")
	}
}

// 整數轉正
func AbsInt32(n int32) int32 {
	if n < 0 {
		return -n
	}
	return n
}

// 调试信息输出
func FmtLog(args ...interface{}) {
	for _, v := range args {
		fmt.Print(v, " ")
	}
	fmt.Print("\n")
}

// 檢查silce中有沒有此元素(int32)
func SearchSliInt(slice []int32, elem int32) bool {
	for _, v := range slice {
		if v == elem {
			return true
		}
	}
	return false
}

// 檢查silce中有沒有此元素(float64)
func SearchSliFlt(slice []float64, elem float64) bool {
	for _, v := range slice {
		if v == elem {
			return true
		}
	}
	return false
}

// 檢查silce中有沒有此元素(str)
func SearchSliStr(slice []string, elem string) bool {
	for _, v := range slice {
		if v == elem {
			return true
		}
	}
	return false
}

// 移除slice中的某個元素(int)
func RemoveSliInt(slice []int32, elem int32) []int32 {
	if len(slice) == 0 {
		return slice
	}
	for i, v := range slice {
		if v == elem {
			slice = append(slice[:i], slice[i+1:]...) // 只删除一个该元素，（唯一的元素）
			// return RemoveEleSlice(slice, elem)// 递归删除全部该元素
			break
		}
	}
	return slice
}

// 移除slice中的某個元素(str)
func RemoveSliStr(slice []string, elem string) []string {
	if len(slice) == 0 {
		return slice
	}
	for i, v := range slice {
		if v == elem {
			slice = append(slice[:i], slice[i+1:]...) // 只删除一个该元素，（唯一的元素）
			// return RemoveEleSlice(slice, elem)// 递归删除全部该元素
			break
		}
	}
	return slice
}

func Str2Int(str string) int {
	num, err := strconv.Atoi(str)
	if err != nil {
		Debug_log("轉int出錯")
	}
	return num
}

func Str2int32(str string) int32 {
	num, err := strconv.ParseInt(str, 10, 32) //轉完可能變int64
	if err != nil {
		Debug_log("轉int32出錯")
	}
	return int32(num)
}

///////HTTP 請求///////

//发送GET请求
//url:请求地址
//response:请求返回的内容
func Get(url string) (response string) {
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("Get-err-1 %+v", err)
		return
		// panic(error)
	}
	defer resp.Body.Close() //必須調用否則可能產生記憶體洩漏

	var buffer [512]byte
	result := bytes.NewBuffer(nil)
	for {
		n, err := resp.Body.Read(buffer[0:])
		result.Write(buffer[0:n])
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			fmt.Printf("Get-err-2 %+v", err)
			// panic(err)
			return
		}
	}

	response = result.String()
	return
}

// int32 slice 加總
func Int32Sum(int32arr []int32) int32 {
	var sum int32 = 0
	for _, i := range int32arr {
		sum += int32(i)
	}
	return sum
}

func Int32SumParallel(numbers []int32) int32 {
	nNum := len(numbers)
	var total int32 = 0
	if nNum < 100000 {
		total = Int32Sum(numbers)
	} else {
		nCPU := runtime.NumCPU()

		ch := make(chan int32)
		for i := 0; i < nCPU; i++ {
			from := i * nNum / nCPU
			to := (i + 1) * nNum / nCPU
			go func() { ch <- Int32Sum(numbers[from:to]) }()
		}
		for i := 0; i < nCPU; i++ {
			total += <-ch
		}
	}

	return total
}

//发送POST请求
//url:请求地址，data:POST请求提交的数据,contentType:请求体格式，如：application/json
//content:请求放回的内容
func Post(url string, data interface{}, contentType string) (content string) {
	jsonStr, _ := json.Marshal(data)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Add("content-type", contentType)
	if err != nil {
		fmt.Printf("Post-Err-1 %+v", err)
		return
	}
	defer req.Body.Close() //必須調用否則可能產生記憶體洩漏

	client := &http.Client{Timeout: 5 * time.Second}
	resp, error := client.Do(req)
	if error != nil {
		fmt.Printf("Post-Err-2 %+v", error)
		return
	}
	defer resp.Body.Close()

	result, _ := ioutil.ReadAll(resp.Body)
	content = string(result)

	return

}

// 氣泡排序Sli(小到大)
func BubbleSort(list []int32) []int32 {
	var temp int32
	var i int
	var j int
	for i = 0; i < len(list)-1; i++ {
		for j = len(list) - 1; j > i; j-- {
			if list[j-1] > list[j] {
				temp = list[j-1]
				list[j-1] = list[j]
				list[j] = temp
			}
		}
	}
	return list
}

//region 快速排序
func division(list []int32, left int32, right int32) int32 {

	// 以最左边的数(left)为基准
	var base int32 = list[left]
	for left < right {
		// 从序列右端开始，向左遍历，直到找到小于base的数
		for left < right && list[right] >= base {
			right--
		}
		// 找到了比base小的元素，将这个元素放到最左边的位置
		list[left] = list[right]
		// 从序列左端开始，向右遍历，直到找到大于base的数
		for left < right && list[left] <= base {
			left++
		}
		// 找到了比base大的元素，将这个元素放到最右边的位置
		list[right] = list[left]

	}
	// 最后将base放到left位置。此时，left位置的左侧数值应该都比left小
	// 而left位置的右侧数值应该都比left大。
	list[left] = base //此时left == right
	//fmt.Println("DONE: base:", base, "\tleft:", left, "\tright:", right)
	return left
}

func QuickSort(list []int32, left int32, right int32) {
	// 左下标一定小于右下标，否则就越界了
	if left < right {
		//对数组进行分割，取出下次分割的基准标号
		var base int32 = division(list, left, right)
		//对“基准标号“左侧的一组数值进行递归的切割，以至于将这些数值完整的排序
		QuickSort(list, left, base-1)
		//对“基准标号“右侧的一组数值进行递归的切割，以至于将这些数值完整的排序
		QuickSort(list, base+1, right)
	}

}

func Removeduplicate(a interface{}) (ret []interface{}) {
	va := reflect.ValueOf(a)
	for i := 0; i < va.Len(); i++ {
		if i > 0 && reflect.DeepEqual(va.Index(i-1).Interface(), va.Index(i).Interface()) {
			continue
		}
		ret = append(ret, va.Index(i).Interface())
	}
	return ret
}

// 去除重複資料 壓測效能優化
func RemoveduplicateMap(DataArr []int32) []int32 {
	var resultArr []int32
	resultmap := make(map[int32]bool, 3)
	for _, i := range DataArr { //大量或者string可以改用i++減少每次賦值的操作
		if _, ok := resultmap[i]; ok {
			continue
		} else {
			resultmap[i] = true
			resultArr = append(resultArr, i)
		}
	}
	// common.Debug_log("單骰結果:%v", SingleDicArr)
	return resultArr
}

// 兩個arr比較是否相等
func IntArrEq(a, b []int) bool {
	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// objectID To Date
func DateFromObjectID(objectID bson.ObjectId) string {
	return objectID.Time().Format(DateLayout)
}
