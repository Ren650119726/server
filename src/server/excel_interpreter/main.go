package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/astaxie/beego"
	"io/ioutil"
	"regexp"
	"root/core/log"
	"strconv"
)

const sheet = "s1"
const sheet2 = "ConvertTable"

type jsonMap map[string]interface{}

func main() {
	beego.LoadAppConfig("ini", "D:/project/repository/bin/app.conf")
	input := beego.AppConfig.DefaultString("CONF::Input_path", "./")
	output := beego.AppConfig.DefaultString("CONF::Output_path", "nill/")
	dir_list, e := ioutil.ReadDir(input)
	if e != nil {
		fmt.Println("read dir error")
		return
	}
	fmt.Println("输出目录:", output)
	for _, filename := range dir_list {
		ret := regexp.MustCompile(`xlsx`).FindStringIndex(filename.Name())
		if ret != nil {
			name := input + "\\" + filename.Name()
			outjson := output + "\\" + filename.Name()[:ret[0]] + "json"
			fmt.Println("解析文件: ", name)
			transform2json(name, outjson)
			//if ret := mashal2Json(name, outjson); ret == "" || ret == "{}" {
			//			//	mashal2Json_row(name, outjson)
			//			//}
		}
	}

}

func transform2json(dirFile string, out string) {
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

	line := "A"
	dataRow := 5 // 数据段从第5行开始
	for {
		// 获得第一行的ID

		pos := fmt.Sprintf("%v%v", line, dataRow)
		key := xls.GetCellValue(sheet, pos)
		if key == "" {
			break
		}
		jsonmap[key] = jsonMap{}
		for {
			// 解析数据行
			line = next(line, len(line)-1)
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

func mashal2Json(dirFile string, out string) string {
	xls, err := excelize.OpenFile(dirFile)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	jsonmap := map[string]jsonMap{}
	orginal_raw := 5

	for ; ; orginal_raw++ {
		pos := fmt.Sprintf("A%v", orginal_raw)
		key := xls.GetCellValue(sheet, pos)
		if key == "" {
			break
		}
		jsonmap[key] = jsonMap{}

		row := 'B'

		for {
			pos2 := fmt.Sprintf("%v2", string(row))
			key_name := xls.GetCellValue(sheet, pos2)
			if key_name == "" {
				break
			}

			cell := xls.GetCellValue(sheet, fmt.Sprintf("%v%v", string(row), orginal_raw))
			typ := xls.GetCellValue(sheet, fmt.Sprintf("%v3", string(row)))
			switch typ {
			case "int":
				icell, err := strconv.Atoi(cell)
				if err == nil {
					jsonmap[key][key_name] = icell
				} else {
					log.Errorf("字段:%v 类型不是int %v", key_name, cell)
				}
			default:
				jsonmap[key][key_name] = cell
			}

			row = rune(int(row) + 1)
		}
	}

	b, _ := json.MarshalIndent(jsonmap, "", "\t")
	ioutil.WriteFile(out, b, 0666)

	return string(b)
}

func mashal2Json_row(dirFile string, out string) {
	xls, err := excelize.OpenFile(dirFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	jsonmap := jsonMap{}
	orginal_raw := 2
	nil_count := 0
	for ; ; orginal_raw++ {
		if nil_count >= 50 {
			break
		}
		pos := fmt.Sprintf("B%v", orginal_raw)
		key := xls.GetCellValue(sheet2, pos)
		if key == "" {
			nil_count++
			continue
		}

		pos = fmt.Sprintf("E%v", orginal_raw)
		val := xls.GetCellValue(sheet2, pos)
		jsonmap[key] = val
		nil_count = 0
	}

	b, _ := json.MarshalIndent(jsonmap, "", "\t")
	ioutil.WriteFile(out, b, 0666)
}
