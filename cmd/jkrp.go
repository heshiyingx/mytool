/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)
var dir string
var ws sync.WaitGroup
// jkrpCmd represents the jkrp command
var jkrpCmd = &cobra.Command{
	Use:   "jkrp",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if !filepath.IsAbs(dir){
			dir,_ = filepath.Abs(dir)
		}

		if _, err := os.Stat(dir);err!=nil{
			panic("该路径不存在")
		}
		findBaseDirInPath(dir)
		fmt.Println("正在处理中")
		ws.Wait()
		fmt.Println("处理完成")
	},
}

func init() {
	rootCmd.AddCommand(jkrpCmd)
	jkrpCmd.PersistentFlags().StringVarP(&dir,"dir","d",".","-p=document path")
}
func findBaseDirInPath(dirPath string) {

	if _, isdir := fileIsExistsAndIsDir(dirPath); isdir {
		fileInfos, _ := ioutil.ReadDir(dirPath)
		for _, v := range fileInfos {
			//fmt.Printf("%d:%v\n", 0, v.Name())
			if v.IsDir() {
				findBaseDirInPath(path.Join(dirPath, v.Name()))
			} else if strings.HasSuffix(v.Name(), ".html") {
				ws.Add(1)
				go findHtmlFileBaseDir(path.Join(dirPath, v.Name()))
			} else if strings.HasSuffix(v.Name(), ".htmlt") {
				os.Remove(path.Join(dirPath, v.Name()))
			}
		}
	}
}

func findHtmlFileBaseDir(filePath string) {
	defer ws.Done()
	re, _ := regexp.Compile(`<base href="([^"]+)"`)
	noselectReg,_ := regexp.Compile("user-select: none;")
	if b, err := ioutil.ReadFile(filePath); err == nil {
        //s:= string(b)
		s := re.ReplaceAllString(string(b), `<base href="."`)
		s = noselectReg.ReplaceAllString(s, "")

		audiTepl := `<audio controls="controls">
  <source src="%s%s" type="audio/mp3" />
</audio>`
		fileName := path.Base(filePath)
		dirPath := path.Dir(filePath)
		nameRe, _ := regexp.Compile(".html")
		mp3BaseName := nameRe.ReplaceAllString(fileName, ".mp3")
		if isExists, _ := fileIsExistsAndIsDir(path.Join(dirPath, mp3BaseName)); isExists {
			audi := fmt.Sprintf(audiTepl, "./", mp3BaseName)
			//fmt.Println(audi)

			contentRe, _ := regexp.Compile(`<audio.+</audio>`)
			s = contentRe.ReplaceAllString(s, audi)
			//
			if len(s) < 1000 {
				fmt.Println(filePath)
			} else {
				//ioutil.WriteFile(filePath+"t", []byte(s), 777)
				//fmt.Println(len(s))
			}

		} else {
			mp3BaseName = nameRe.ReplaceAllString(fileName, ".m4a")
			if isExists, _ := fileIsExistsAndIsDir(path.Join(dirPath, mp3BaseName)); isExists {
				audi := fmt.Sprintf(audiTepl, "./", mp3BaseName)
				//fmt.Println(audi)

				contentRe, _ := regexp.Compile(`<audio.+</audio>`)
				s = contentRe.ReplaceAllString(s, audi)
				//
				if len(s) < 1000 {
					fmt.Println(filePath)
				} else {
					//ioutil.WriteFile(filePath+"t", []byte(s), 777)
					//fmt.Println(len(s))
				}

			}
		}
		if len(s) < 1000 {
			fmt.Println(filePath)
		} else {

			ioutil.WriteFile(filePath, []byte(s), 777)
		}
		//

	} else {
		fmt.Sprintf("%v 更改dir失败\n", filePath)
	}
}

func findHtmlAudio(filePath string) {
	audiTepl := `<audio controls="controls">
  <source src="%s%s" type="audio/mp3" />
</audio>`
	fileName := path.Base(filePath)
	dirPath := path.Dir(filePath)
	nameRe, _ := regexp.Compile(".html")
	mp3BaseName := nameRe.ReplaceAllString(fileName, ".mp3")
	if isExists, _ := fileIsExistsAndIsDir(path.Join(dirPath, mp3BaseName)); isExists {
		audi := fmt.Sprintf(audiTepl, "./", mp3BaseName)
		fmt.Println(audi)

		contentRe, _ := regexp.Compile(`<audio.+</audio>`)
		if b, err := ioutil.ReadFile(filePath); err == nil {
			s := contentRe.ReplaceAllString(string(b), audi)
			ioutil.WriteFile(filePath+"t", []byte(s), 777)
		}
	}

}

func fileIsExistsAndIsDir(path string) (isExists bool, isDir bool) {
	f, err := os.Stat(path)
	if err != nil {
		return false, false
	}
	if f.IsDir() {
		return true, true
	}
	return true, false

}
