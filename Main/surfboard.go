package main

import (
	"bufio"
	"flag"
	"log"
	"net/http"
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

	releasePtr := flag.Bool("notes", false, "used to switch on Release Note production")
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

	dependancies, notes := generateMap(projectpath, *releasePtr)

	top := join("<html><head><link rel='stylesheet' href='chart-tbl.css'></head><body><h1>Surfboard for ", project, " Helm Chart</h1> <table>")
	tail := "</body></html>"

	output := join(top, dependancies, "</table>", notes, tail)
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

func generateMap(projectpath string, releasenotes bool) (string, string) {
	depth := 0
	dependancies := ""
	notes := ""

	visit := func(path string, info os.FileInfo, err error) error {

		if info.IsDir() {
			depth, dependancies = processDir(path, depth, dependancies)
		} else if info.Name() == "_release_notes.yaml" && releasenotes {
			note := generateReleaseNotes(path)
			notes = join(notes, note)
		}
		return nil
	}

	err := filepath.Walk(projectpath, visit)
	if err != nil {
		log.Fatal(err)
	}
	return dependancies, notes
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
	entry := join("<tr>", pad, "<td><h2>", "<a href='#", config.name, "'>", config.name, "</a>", "</h2>", config.version, "</br>", config.description, "</td></tr>")
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

	// Could get issue summary if logged in
	// https://jira.ipttools.info/rest/api/latest/issue/EE-15620?fields=summary
	//curl "https://<user>:<password>@jira.ipttools.info/rest/api/latest/issue/EE-15620?fields=summary"

	println("Generate Release Notes")
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	parts := strings.Split(filename, "/")
	n := len(parts)
	service := parts[n-2]
	println(service)

	output := join("<h2>Release Notes for  <a name='", service, "'>", service, "</a> </h2>")
	scanner := bufio.NewScanner(file)
	issuepattern, _ := regexp.Compile("[A-Z]{2,}-[0-9]{4,}")
	versionpattern, _ := regexp.Compile("\\.[0-9]{0,3}-")
	keys := make(map[string]bool)
	for scanner.Scan() {
		entry := scanner.Text()
		if versionpattern.MatchString(entry) {
			if !strings.Contains(entry, "JENKINS") {
				keys = make(map[string]bool)
				output = join(output, entry, "</br>")
			}
		} else {
			entry := issuepattern.FindString(scanner.Text())
			if _, value := keys[entry]; !value {
				keys[entry] = true
				output = join(output, "&nbsp;&nbsp;&nbsp<a href='https://jira.ipttools.info/browse/", entry, "' target='_blank'>", entry, "</a>", "</br>")
			}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return output
}

func unique(sSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range sSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func getJiraSummary(user string, password string, ticket string) {
	url := join("'https://", user, ":", password, "@jira.ipttools.info/rest/api/latest/issue/", ticket, "?fields=summary'")
	resp, err := http.Get(url)
	if err != nil {
		// handle err
	}
	defer resp.Body.Close()
}
