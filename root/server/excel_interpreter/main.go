package main

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/astaxie/beego"
	"io/ioutil"
	"os"
	"regexp"
	"root/core/utils"
	"strconv"
	"strings"
	"time"
)

const sheet = "s1"
type jsonMap map[string]interface{}

func main() {
	beego.LoadAppConfig("ini", "./app.conf")
	inputstr := beego.AppConfig.DefaultString("CONF::Input_path", "./")
	outputstr := beego.AppConfig.DefaultString("CONF::Output_path", "C:/Users/wwj/Desktop/server/root/config/")
	fmt.Println("输出目录:", outputstr)
	inputs := strings.Split(inputstr, " ")
	for _,v := range inputs{
		fmt.Println("输入目录:", v)
	}
	time := time.Now().Format(utils.STD_NUMBER_FORMAT)

	// 打包并删除旧文件
	if !compressZIP(outputstr,time){
		return
	}

	// 生成新文件
	for _,input := range inputs{
		dir_list, e := ioutil.ReadDir(input)
		if e != nil {
			fmt.Println("read dir error")
			return
		}

		for _, filename := range dir_list {
			ret := regexp.MustCompile(`xlsx`).FindStringIndex(filename.Name())
			if ret != nil {
				name := input + "/" + filename.Name()
				jsonname := filename.Name()[:ret[0]-1] +"_"+ time + ".json"
				outjson := outputstr + "/" + jsonname
				transform2json(name, outjson)
				fmt.Printf("解析文件:%-70v %-30v\n",name,jsonname)
			} else {
				//fmt.Println("igrone:",filename.Name())
				continue
			}
		}
	}
}

func transform2json(dirFile string, out string){
	xls, err := excelize.OpenFile(dirFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	jsonmap := map[string]jsonMap{}
	var next func(str string, len int) string
	next = func(str string, len int) string {
		end := len - 1
		c := []rune(str)
		if c[end] == 'Z' {
			c[end] = 'A'
			if end == 0 {
				c = append([]rune("A"), c...)
				return string(c)
			} else {
				return next(string(c), len-1)
			}
		} else {
			c[end] = c[end] + 1
			return string(c)
		}
		return string(c)
	}


	dataRow := 5 // 数据段从第5行开始
	for {
		// 获得第一行的ID
		line := "A"
		pos := fmt.Sprintf("%v%v", line, dataRow)
		key := xls.GetCellValue(sheet, pos)
		if key == "" {
			break
		}
		jsonmap[key] = jsonMap{}
		for {
			// 解析数据行
			line = next(line, len(line))
			name_pos := fmt.Sprintf("%v%v", line, 2)
			type_pos := fmt.Sprintf("%v%v", line, 3)
			usd_pos := fmt.Sprintf("%v%v", line, 4)
			val_pos := fmt.Sprintf("%v%v", line, dataRow)

			n := xls.GetCellValue(sheet, name_pos)
			t := xls.GetCellValue(sheet, type_pos)
			v := xls.GetCellValue(sheet, val_pos)
			used := xls.GetCellValue(sheet, usd_pos)
			if n == "" {
				break
			}
			if used == "p" {
				continue
			}
			if t == "int" {
				iv, e := strconv.Atoi(v)
				if e != nil {
					e := errors.New(fmt.Sprint("行:[%v] 列:[%v] 值:[%v] 类型不是int 请检查", dataRow, line, v))
					panic(e)
				}
				jsonmap[key][n] = iv
			} else if t == "string" {
				jsonmap[key][n] = v
			}
		}
		dataRow++
	}

	b, _ := json.MarshalIndent(jsonmap, "", "\t")
	ioutil.WriteFile(out, b, 0666)
	return
}

func compressZIP(outputPath,time string) bool {
	// 先将旧文件打包保存，在生成新文件
	dir_list, e := ioutil.ReadDir(outputPath)
	if e != nil {
		fmt.Println("read dir error")
		return false
	}
	for i := len(dir_list)-1;i >=0;i--{
		if !strings.Contains(dir_list[i].Name(), "json") {
			dir_list = append(dir_list[:i],dir_list[i+1:]...)
		}
	}
	if len(dir_list) == 0{
		return true
	}
	// 打包旧文件
	file_zip := outputPath +fmt.Sprintf("config_%v.zip",time)
	fzip, e := os.Create(file_zip)
	if e != nil {
		fmt.Println("文件打开失败:",file_zip)
		return false
	}
	w := zip.NewWriter(fzip)

	defer func() {
		w.Close()
		fzip.Close()

	}()

	for _,file := range dir_list{
		if strings.Contains(file.Name(), "json") {
			fw, err := w.Create(file.Name())
			if err != nil {
				fmt.Println(err)
				return false
			}
			filecontent, err := ioutil.ReadFile(outputPath + file.Name())
			if err != nil {
				fmt.Println(err)
				return false
			}
			_, err = fw.Write(filecontent)
			if err != nil {
				fmt.Println(err)
				return false
			}
			fmt.Println("打包并移除文件 ", file.Name())
			os.Remove(outputPath + file.Name())
		}
	}
	return true
}