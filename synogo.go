package main

/*
TODO
 - download station info
 - clear finished tasks
 - handle goroutine errors
*/

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/macpoint/synogo/services"
	"github.com/macpoint/synogo/synoclient"
	"github.com/olekukonko/tablewriter"
)

func main() {

	config, err := synoclient.LoadJsonConfiguration(filepath.Join(os.Getenv("HOME"), ".synogo.json"))
	if err != nil {
		fmt.Println(err)
		return
	}

	client := &synoclient.Client{
		Host:     config.Host,
		Scheme:   config.Scheme,
		Username: config.Username,
		Password: config.Password,
		Session:  "DownloadStation",
		Timeout:  config.Timeout,
	}

	file := flag.String("f", "", "Create download task from file")
	url := flag.String("u", "", "Create download task from url")
	list := flag.Bool("l", false, "List existing download tasks")
	delete := flag.String("d", "", "Delete tasks ids separated by comma")
	pause := flag.String("p", "", "Pause tasks ids separated by comma")
	resume := flag.String("r", "", "Resume tasks ids separated by comma")

	flag.Parse()
	if *file != "" {
		createDownloadTaskFromFile(client, *file)
		return
	}
	if *url != "" {
		createDownloadTaskfromUrl(client, *url)
		return
	}

	if *list {
		getDownloadTasks(client)
		return
	}

	if *delete != "" {
		deleteDownloadTasks(client, *delete)
		return
	}

	if *pause != "" {
		pauseDownloadTasks(client, *pause)
		return
	}

	if *resume != "" {
		resumeDownloadTasks(client, *resume)
		return
	}

	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
}

func deleteDownloadTasks(client *synoclient.Client, tasks string) {
	// Login
	_, err := client.Login()
	if err != nil {
		fmt.Println(err)
		return
	}

	resp, err := services.DeleteDownloadStationTasks(client, tasks)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 'delete' response always returns success: true
	// response is an array of response objects with following parameters
	/*
		"data": [
			{
				"error": 544,
				"id": "dbid_264"
			},
			{
				"error": 0,
				"id": "dbid_267"
			}
		],
	*/
	results := client.GetData(resp).([]interface{})
	for _, result := range results {
		r := result.(map[string]interface{})
		if int(r["error"].(float64)) > 0 {
			fmt.Printf("Could not delete task id %v (%v).\n", r["id"], r["error"])
		} else {
			fmt.Printf("Task %v deleted.\n", r["id"])
		}
	}

	// Logout
	client.Logout()
}

func resumeDownloadTasks(client *synoclient.Client, tasks string) {
	// Login
	_, err := client.Login()
	if err != nil {
		fmt.Println(err)
		return
	}

	resp, err := services.ResumeDownloadStationTasks(client, tasks)
	if err != nil {
		fmt.Println(err)
		return
	}

	results := client.GetData(resp).([]interface{})
	for _, result := range results {
		r := result.(map[string]interface{})
		if int(r["error"].(float64)) > 0 {
			fmt.Printf("Could not resume task id %v (%v).\n", r["id"], r["error"])
		} else {
			fmt.Printf("Task %v resumed.\n", r["id"])
		}
	}

	// Logout
	client.Logout()
}

func pauseDownloadTasks(client *synoclient.Client, tasks string) {
	// Login
	_, err := client.Login()
	if err != nil {
		fmt.Println(err)
		return
	}

	resp, err := services.PauseDownloadStationTasks(client, tasks)
	if err != nil {
		fmt.Println(err)
		return
	}

	results := client.GetData(resp).([]interface{})
	for _, result := range results {
		r := result.(map[string]interface{})
		if int(r["error"].(float64)) > 0 {
			fmt.Printf("Could not pause task id %v (%v).\n", r["id"], r["error"])
		} else {
			fmt.Printf("Task %v paused.\n", r["id"])
		}
	}

	// Logout
	client.Logout()
}

func createDownloadTaskFromFile(client *synoclient.Client, filepath string) {
	// Login
	_, err := client.Login()
	if err != nil {
		fmt.Println(err)
		return
	}

	file, err := os.Open(filepath)
	if err != nil {
		fmt.Printf("Could not open file %v\n", filepath)
		os.Exit(1)
	}
	defer file.Close()

	// create sync & queue & add workers to sync group
	var wg sync.WaitGroup
	const noOfWorkers = 3

	wg.Add(noOfWorkers)
	fileProcessQueue := make(chan string, 5)

	// create workers
	for gr := 1; gr <= noOfWorkers; gr++ {
		go services.CreateDownloadStationTask(client, fileProcessQueue, &wg)
	}

	// fill the queue with each line of the file
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fileProcessQueue <- scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	close(fileProcessQueue)
	wg.Wait()
	client.Logout()
}

func createDownloadTaskfromUrl(client *synoclient.Client, url string) {
	// Login
	_, err := client.Login()
	if err != nil {
		fmt.Println(err)
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)
	urlProcessQueue := make(chan string)
	go services.CreateDownloadStationTask(client, urlProcessQueue, &wg)
	urlProcessQueue <- url

	close(urlProcessQueue)
	wg.Wait()
	client.Logout()
}

func getDownloadTasks(client *synoclient.Client) {

	// Login
	_, err := client.Login()
	if err != nil {
		fmt.Println(err)
		return
	}

	downloadTasks, e := services.GetDownloadStationTasks(client, "all")
	if e != nil {
		fmt.Println(e)
	}

	if len(downloadTasks) > 0 {
		formatDownloadTasks(downloadTasks)
	} else {
		fmt.Println("No download tasks found.")
	}

	// Logout
	client.Logout()
}

func formatDownloadTasks(dstasks []services.DownloadStationTask) {
	var data [][]string
	for _, task := range dstasks {
		var downloaded int64
		if task.Size != 0 {
			downloaded = task.AdditinalTaskInfo.TaskTransfer.SizeDownloaded / (task.Size / 100)
		} else {
			downloaded = 0
		}
		data = append(data, []string{
			task.ID,
			task.Title,
			//strconv.FormatInt(task.Size, 10),
			ByteCountSI(task.Size),
			task.Type,
			task.Status,
			fmt.Sprintf("%v%%", strconv.FormatInt(downloaded, 10)),
			task.AdditinalTaskInfo.TaskDetail.Destination,
		})
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Id", "Title", "Size", "Type", "Status", "Downloaded", "Destination"})

	for _, v := range data {
		table.Append(v)
	}
	table.Render()
}

// ByteCountSI ... https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format/
func ByteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}
