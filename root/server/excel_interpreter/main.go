package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/astaxie/beego"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
)

const sheet = "s1"
type jsonMap map[string]interface{}

func main() {
	beego.LoadAppConfig("ini", "./app.conf")
	inputstr := beego.AppConfig.DefaultString("CONF::Input_path", "./")
	output := beego.AppConfig.DefaultString("CONF::Output_path", "nill/")
	fmt.Println("输出目录:", output)
	inputs := strings.Split(inputstr, " ")
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
				jsonname := filename.Name()[:ret[0]] + "json"
				outjson := output + "/" + jsonname
				transform2json(name, outjson)
				fmt.Printf("解析文件:%-70v %-30v\n",name,jsonname)
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