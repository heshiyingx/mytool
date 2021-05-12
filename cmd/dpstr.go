package cmd

import (
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type dpStrInfo struct {
	Path      string
	Recursive bool
	DelStr    string
}

var info dpStrInfo

func init() {
	rootCmd.AddCommand(dpstrCmd)
	dpstrCmd.PersistentFlags().StringVarP(&info.Path, "path", "p", ".", "文件路径")
	dpstrCmd.PersistentFlags().StringVarP(&info.DelStr, "delStr", "d", "", "删除的字符串")
	dpstrCmd.PersistentFlags().BoolVarP(&info.Recursive, "recursive", "R", false, "是否递归")

}

var dpstrCmd = &cobra.Command{
	Use:   "dpstr",
	Short: "delete part string in fileName",
	Long: `mysql table to struct
eg:dpstr -d="删除的字符串" -p="文件路径" -R=true`,
	Run: func(cmd *cobra.Command, args []string) {
		if info.DelStr==""{
			log.Println("删除的字段不能为空")
			return
		}
		err := renameFile(info.Path, info.DelStr, info.Recursive, nil)
		if err != nil {
			log.Println(err)
		}
	},
}
type finishFn  func()
func renameFile(filePathSrc string, delStr string,isRecursive bool,fn finishFn) error {
	defer func() {
		if fn!=nil{
			fn()
		}
	}()
	var wg sync.WaitGroup
	//	1.转成绝对路径
	absPath, err := filepath.Abs(filePathSrc)
	if err != nil {
		return err
	}
	//log.Println("path:",absPath)
	stat, err := os.Stat(absPath)
	if err != nil {
		return err
	}
	// 2.判断是否递归，如果是递归并且是文件夹递归调用此方法
	if isRecursive&&stat.IsDir(){
		subFiles, err := os.ReadDir(absPath)
		if err != nil {
			return err
		}
		for _,f := range subFiles{
			wg.Add(1)
			go renameFile(filepath.Join(absPath,f.Name()),delStr,isRecursive, func() {
				wg.Done()
			})
			wg.Wait()
		}
	}
	// do
	destName := getDestName(filePathSrc, delStr)
	return os.Rename(filePathSrc, destName)
}

func getDestName(absPath string,deleStr string)string  {
	name := filepath.Base(absPath)
	dir := filepath.Dir(absPath)
	dstName := strings.Replace(name, deleStr, "", -1)
	log.Println("new：",dstName,"\nold:",name)
	return filepath.Join(dir,dstName)
}
