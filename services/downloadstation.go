package services

import (
	"fmt"
	"sync"

	"github.com/macpoint/synogo/synoclient"
)

type DownloadStationTask struct {
	ID                string
	Type              string
	Size              int64
	Status            string
	Title             string
	Username          string
	AdditinalTaskInfo AdditinalTaskInfo
}

type AdditinalTaskInfo struct {
	TaskTransfer TaskTransfer
}

type TaskTransfer struct {
	SizeDownloaded int64
	SpeedDownload  int64
}

func GetDownloadStationTasks(client *synoclient.Client, status string) ([]DownloadStationTask, error) {

	params := map[string]string{
		"api":        "SYNO.DownloadStation.Task",
		"version":    "1",
		"method":     "list",
		"additional": "transfer",
	}

	resp, err := client.Get("webapi/DownloadStation/task.cgi", params)
	if err != nil {
		return nil, err
	}

	tasks := client.GetData(resp).(map[string]interface{})["tasks"].([]interface{})
	var downloadTasks []DownloadStationTask
	for _, task := range tasks {
		t := task.(map[string]interface{})
		transferInfo := t["additional"].(map[string]interface{})["transfer"].(map[string]interface{})
		downloadTasks = append(
			downloadTasks,
			DownloadStationTask{
				ID:       t["id"].(string),
				Type:     t["type"].(string),
				Size:     int64(t["size"].(float64)),
				Status:   t["status"].(string),
				Title:    t["title"].(string),
				Username: t["username"].(string),
				AdditinalTaskInfo: AdditinalTaskInfo{
					TaskTransfer: TaskTransfer{
						SizeDownloaded: int64(transferInfo["size_downloaded"].(float64)),
						SpeedDownload:  int64(transferInfo["speed_download"].(float64)),
					},
				},
			})
	}

	return downloadTasks, nil

}

func CreateDownloadStationTask(client *synoclient.Client, fileQueue <-chan string, wg *sync.WaitGroup) error {

	params := map[string]string{
		"api":     "SYNO.DownloadStation.Task",
		"version": "1",
		"method":  "create",
	}
	defer wg.Done()
	for filename := range fileQueue {
		params["uri"] = filename
		fmt.Printf("Adding %v ...\n", truncateString(filename, 70))
		_, err := client.Get("webapi/DownloadStation/task.cgi", params)
		if err != nil {
			return err
		}
	}
	return nil
}

func DeleteDownloadStationTasks(client *synoclient.Client, taskIds string) (response string, err error) {
	params := map[string]string{
		"api":     "SYNO.DownloadStation.Task",
		"version": "1",
		"method":  "delete",
		"id":      taskIds,
	}
	response, err = client.Get("webapi/DownloadStation/task.cgi", params)
	if err != nil {
		return "", err
	}
	return response, nil
}

func PauseDownloadStationTasks(client *synoclient.Client, taskIds string) (response string, err error) {
	params := map[string]string{
		"api":     "SYNO.DownloadStation.Task",
		"version": "1",
		"method":  "pause",
		"id":      taskIds,
	}
	response, err = client.Get("webapi/DownloadStation/task.cgi", params)
	if err != nil {
		return "", err
	}
	return response, nil
}

func ResumeDownloadStationTasks(client *synoclient.Client, taskIds string) (response string, err error) {
	params := map[string]string{
		"api":     "SYNO.DownloadStation.Task",
		"version": "1",
		"method":  "resume",
		"id":      taskIds,
	}
	response, err = client.Get("webapi/DownloadStation/task.cgi", params)
	if err != nil {
		return "", err
	}
	return response, nil
}

func truncateString(str string, num int) string {
	truncated := str
	if len(str) > num {
		if num > 3 {
			num -= 3
		}
		truncated = str[0:num] + "..."
	}
	return truncated
}
