package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

	releasePtr := flag.Int("notes", 0, "used to switch on Release Note production")
	flag.Parse()

	project := flag.Arg(0)
	ver := flag.Arg(1)

	println(project)
	println(ver)

	getTheChart(project, ver)

	projectpath := join("./", project)

	//	notes := generateReleaseNotes(join(projectpath, "/_release_notes.yaml"))
	//	output = join(output, notes)
	//}

	output := generateMap(projectpath, *releasePtr)

	top := join("<html><head><link rel='stylesheet' href='chart-tbl.css'></head><body><h1>Surfboard for ", project, " Helm Chart</h1> <table>")
	tail := "</body></html>"

	output = join(top, output, tail)
	writeMap(output, project)

	os.RemoveAll(projectpath)
}

func getTheChart(project string, ver string) {
	cmd := exec.Command("helm", "fetch", "--untar", join("chartmuseum/", project), "--version", ver)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}

func generateMap(projectpath string, notesDepth int) string {
	depth := 0
	output := ""

	visit := func(path string, info os.FileInfo, err error) error {

		if info.IsDir() {
			depth, output = processDir(path, depth, output)
		} else if info.Name() == "_release_notes.yaml" && depth <= notesDepth {
			notes := generateReleaseNotes(path)
			output = join(output, notes)
		}
		return nil
	}

	err := filepath.Walk(projectpath, visit)
	if err != nil {
		log.Fatal(err)
	}
	return output
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
		output = join(output, addEntry(data, depth))
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

func addEntry(config config, depth int) string {
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

func writeMap(output string, project string) {
	filename := join("./", project, "-map.html")
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}

	f.WriteString(output)
	f.Sync()

	cmd := exec.Command("open", filename)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
	println("map created")
}

func generateReleaseNotes(filename string) string {
	println("Generate Release Notes")
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	output := join("<h2>Release Notes for ", filename, "</h2>")
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		output = join(output, scanner.Text(), "</br>")
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	regxpattern, _ := regexp.Compile(" [0-9a-f]*:")
	output = regxpattern.ReplaceAllString(output, " ")
	return output
}
