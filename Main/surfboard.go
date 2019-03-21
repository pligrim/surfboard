package main

import (
	"bufio"
	"flag"
	"io/ioutil"
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

type rulesStruc struct {
	Target      string
	Path        string
	ServiceName string
	ServicePort string
}

func main() {

	releasePtr := flag.Bool("notes", false, "used to switch on Release Note production")
	userPtr := flag.String("user", "", "User name for jira user to collect ticket summary")
	passPtr := flag.String("token", "", "Token for jira user to collect ticket summary")
	routesPtr := flag.Bool("routes", false, "if true will show version routing")

	flag.Parse()

	path := flag.Arg(0)
	ver := flag.Arg(1)

	project := strings.Split(path, "/")[1]

	println(project)
	println(ver)

	getTheChart(path, ver)

	projectpath := join("./", project)

	dependancies, notes, routes := generateMap(projectpath, *releasePtr, *routesPtr, *userPtr, *passPtr)

	top := join("<html><head><link rel='stylesheet' href='chart-tbl.css'></head><body><h1>Surfboard for ", project, " Helm Chart</h1> <table>")
	tail := "</body></html>"

	output := join(top, dependancies, "</table>", notes, routes, tail)
	writeMap(output, project)

	os.RemoveAll(projectpath)
}

func getTheChart(path string, ver string) {
	cmd := exec.Command("helm", "fetch", "--untar", path, "--version", ver)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}

func generateMap(projectpath string, releasenotes bool, versionroutes bool, user string, password string) (string, string, string) {
	depth := 0
	dependancies := ""
	notes := ""
	routing := ""

	visit := func(path string, info os.FileInfo, err error) error {

		if info.IsDir() {
			depth, dependancies = processDir(path, depth, dependancies)
		} else if strings.Contains(info.Name(), "-values.yaml") && versionroutes {
			route := routes(path, info.Name())
			routing = join(routing, route)
		} else if info.Name() == "_release_notes.yaml" && releasenotes {
			note := generateReleaseNotes(path, user, password)
			notes = join(notes, note)
		}
		return nil
	}

	err := filepath.Walk(projectpath, visit)
	if err != nil {
		log.Fatal(err)
	}
	return dependancies, notes, routing
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

func generateReleaseNotes(filename string, user string, password string) string {

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
				if user != "" && password != "" {
					getJiraSummary(user, password, entry)
				}
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

func getJiraSummary(user string, token string, ticket string) string {
	url := join("https://jira.ipttools.info/rest/api/latest/issue/", ticket, "?fields=summary")

	// Generated by curl-to-Go: https://mholt.github.io/curl-to-go

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// handle err
	}
	req.SetBasicAuth(user, token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	println(string(body))

	return ""
}

func routes(path string, filename string) string {
	fullname := join(path)
	output := ""
	println("")
	print("Generate Version Routes for ")
	println(fullname)

	shortName := strings.Replace(filename, ".yaml", "", 1)

	viper.SetConfigType("yaml")
	viper.SetConfigName(shortName)
	viper.AddConfigPath(strings.Replace(path, filename, "", 1))

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {
		panic(err)
	}

	gway := viper.GetString("status-api-gateway.ingress.host")
	hasRouting := gway != ""

	if hasRouting {

		namespace := strings.Replace(shortName, "-values", "", 1)
		env := strings.Split(namespace, "-")[3]

		output = join("<h2>Namespace:  <a name='", env, "'>", namespace, "</a> </h2>")

		var rules []rulesStruc
		err := viper.UnmarshalKey("status-api-gateway.ingress.rules", &rules)
		if err != nil {
			panic("Unable to unmarshal hosts")
		}
		for _, r := range rules {
			service := strings.Replace(r.ServiceName, "-", " ", -1)
			output = join(output, service, "</br>")
		}

		viper.Reset()
	}

	return output
}
