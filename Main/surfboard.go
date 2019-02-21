package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

type config struct {
	apiVersion  string
	appVersion  string
	description string
	name        string
	version     string
	depth       string
}

func main() {

	project := os.Args[1]
	ver := os.Args[2]

	println(project)
	println(ver)

	cmd := exec.Command("helm", "fetch", "--untar", join("chartmuseum/", project), "--version", ver)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		panic(err)
	}

	dirList := make([]string, 0)

	visit := func(path string, info os.FileInfo, err error) error {

		if info.IsDir() {
			//fmt.Println("dir:  ", path)
			dirList = append(dirList, path)
		} else if info.Name() == "Chart.yaml" {
			//fmt.Println("  file: ", path)
		}

		return nil
	}

	projectpath := join("./", project)
	err = filepath.Walk(projectpath, visit)
	if err != nil {
		log.Fatal(err)
	}

	depth := 0
	output := ""
	for _, path := range dirList {

		depth, output = processDir(path, depth, output)

	}

	top := "<html><head><link rel='stylesheet' href='chart-tbl.css'></head><body><h1>Surfboard</h1><table>"
	tail := "</body></html>"

	println()

	f, err := os.Create("../map.html")
	if err != nil {
		log.Fatal(err)
	}
	output = join(top, output, tail)
	f.WriteString(output)
	f.Sync()

	cmd = exec.Command("open", "../map.html")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
	println("map created")
	os.RemoveAll(projectpath)
}

func processDir(path string, previousDepth int, output string) (int, string) {
	viper.SetConfigType("yaml")
	viper.SetConfigName("chart")
	viper.AddConfigPath(path)

	depth := previousDepth

	err := viper.ReadInConfig() // Find and read the config file
	depth = strings.Count(path, "/")

	if err == nil {

		data := getConfig()
		data.depth = strconv.Itoa(depth)

		output = join(output, addEntry2(data, depth))
	}
	viper.Reset()
	return depth, output
}

func getConfig() config {
	cfig := config{}

	cfig.appVersion = viper.GetString("appVersion")
	cfig.description = viper.GetString("description")
	cfig.name = viper.GetString("name")
	cfig.version = viper.GetString("version")

	return cfig
}

func addEntry2(config config, depth int) string {
	pad := strings.Repeat("<td></td>", depth)
	entry := join("<tr>", pad, "<td><h2>", config.name, "</h2>", config.version, "</br>", config.description, "</td></tr>")
	return entry
}

func join(strs ...string) string {
	var ret string
	for _, str := range strs {
		ret += str
	}
	return ret
}
