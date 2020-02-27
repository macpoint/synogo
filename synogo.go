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
	"os"
	"path/filepath"
	"strconv"
	"sync"

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
	move := flag.String("m", "", "Move downloaded file to destination")

	flag.Parse()
	if *file != "" {
		createDownloadTaskFromFile(client, *file)
		return
	}
	if *url != "" {
		createDownloadTaskfromURL(client, *url)
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

	if *move != "" {

		if len(flag.Args()) != 1 {
			printUsage()
			return
		}
		moveDownloadedFile(client, *move, flag.Args()[0])
		return
	}

	printUsage()

}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
}

func moveDownloadedFile(client *synoclient.Client, taskID string, destination string) {
	// Login
	_, err := client.Login()
	if err != nil {
		fmt.Println(err)
		return
	}

	task, err := client.GetDownloadStationTask(taskID)
	if err != nil {
		fmt.Println(err)
		return
	}

	if task.Status != "finished" {
		fmt.Printf("File %v cannot be moved. It has not beed downloaded yet.\n", task.Title)
		return
	}

	fileToMove := "/" + filepath.Join(task.AdditinalTaskInfo.TaskDetail.Destination, task.Title)
	desiredFileName := filepath.Base(destination)
	renamedFile, err := client.RenameFile(fileToMove, desiredFileName)
	if err != nil {
		fmt.Println(err)
		return
	}

	destinationDir := filepath.Dir(destination)
	err = client.MoveFile(renamedFile, destinationDir)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("File moved.")
	client.Logout()
}

func deleteDownloadTasks(client *synoclient.Client, tasks string) {
	// Login
	_, err := client.Login()
	if err != nil {
		fmt.Println(err)
		return
	}

	resp, err := client.DeleteDownloadStationTasks(tasks)
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
			fmt.Printf("Could not delete task id %v (%v).\n", r["id"], synoclient.DsSynoErrors[int(r["error"].(float64))])
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

	resp, err := client.ResumeDownloadStationTasks(tasks)
	if err != nil {
		fmt.Println(err)
		return
	}

	results := client.GetData(resp).([]interface{})
	for _, result := range results {
		r := result.(map[string]interface{})
		if int(r["error"].(float64)) > 0 {
			fmt.Printf("Could not resume task id %v (%v).\n", r["id"], synoclient.DsSynoErrors[int(r["error"].(float64))])
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

	resp, err := client.PauseDownloadStationTasks(tasks)
	if err != nil {
		fmt.Println(err)
		return
	}

	results := client.GetData(resp).([]interface{})
	for _, result := range results {
		r := result.(map[string]interface{})
		if int(r["error"].(float64)) > 0 {
			fmt.Printf("Could not pause task id %v (%v).\n", r["id"], synoclient.DsSynoErrors[int(r["error"].(float64))])
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
	var processWg sync.WaitGroup
	var errorWg sync.WaitGroup
	const noOfWorkers = 3

	processWg.Add(noOfWorkers)
	errorWg.Add(1)
	fileProcessQueue := make(chan string, noOfWorkers)
	errorQueue := make(chan *synoclient.TaskAddError)

	// create workers
	for gr := 1; gr <= noOfWorkers; gr++ {
		go client.CreateDownloadStationTask(fileProcessQueue, errorQueue, &processWg)
	}

	// read the error queue
	go readErrors(errorQueue, &errorWg)

	// fill the queue with each line of the file
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fileProcessQueue <- scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}

	close(fileProcessQueue)
	processWg.Wait()

	close(errorQueue)
	errorWg.Wait()

	client.Logout()
}

func readErrors(errorQueue <-chan *synoclient.TaskAddError, wg *sync.WaitGroup) {
	defer wg.Done()
	for data := range errorQueue {
		if data.Err != nil {
			fmt.Printf("Task %v not added: %v\n", data.Name, data.Err)
		}
	}
}

func createDownloadTaskfromURL(client *synoclient.Client, url string) {
	// Login
	_, err := client.Login()
	if err != nil {
		fmt.Println(err)
		return
	}

	var processWg sync.WaitGroup
	var errorWg sync.WaitGroup

	processWg.Add(1)
	errorWg.Add(1)

	urlProcessQueue := make(chan string)
	errorQueue := make(chan *synoclient.TaskAddError, 5)

	go client.CreateDownloadStationTask(urlProcessQueue, errorQueue, &processWg)
	go readErrors(errorQueue, &errorWg)

	urlProcessQueue <- url

	close(urlProcessQueue)
	processWg.Wait()

	close(errorQueue)
	errorWg.Wait()

	client.Logout()
}

func getDownloadTasks(client *synoclient.Client) {

	// Login
	_, err := client.Login()
	if err != nil {
		fmt.Println(err)
		return
	}

	downloadTasks, e := client.ListDownloadStationTasks()
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

func formatDownloadTasks(dstasks []synoclient.DownloadStationTask) {
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
