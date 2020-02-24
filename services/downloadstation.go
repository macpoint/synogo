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
	TaskDetail   TaskDetail
}

type TaskTransfer struct {
	SizeDownloaded int64
	SpeedDownload  int64
}

type TaskDetail struct {
	Destination string
	Uri         string
}

type DsError struct {
	desc string
}

func (dsError *DsError) Error() string {
	return fmt.Sprintf("Download Station error: %v", dsError.desc)
}

// GetDownloadStationTask returns one DownloadStationTask
func GetDownloadStationTask(client *synoclient.Client, taskID string) (DownloadStationTask, error) {
	tasksMap, _ := GetDownloadStationTasks(client, taskID)
	var dsTask DownloadStationTask
	for i := range tasksMap {
		if tasksMap[i].ID == taskID {
			dsTask = tasksMap[i]
		}
	}

	if dsTask.ID == "" {
		return dsTask, &DsError{"Task not found."}
	}

	return dsTask, nil
}

// GetDownloadStationTasks returns download tasks using 'getinfo'
func GetDownloadStationTasks(client *synoclient.Client, taskIds string) ([]DownloadStationTask, error) {
	params := map[string]string{
		"api":     "SYNO.DownloadStation.Task",
		"version": "1",
		"method":  "getinfo",
		// SynoAPI accepts multiple IPs separated by comma
		"id":         taskIds,
		"additional": "transfer,detail",
	}

	resp, err := client.Get("webapi/DownloadStation/task.cgi", params)
	if err != nil {
		return nil, err
	}

	tasks := client.GetData(resp).(map[string]interface{})["tasks"].([]interface{})
	downloadTasks := mapDownloadStationTask(tasks)

	return downloadTasks, nil

}

func ListDownloadStationTasks(client *synoclient.Client) ([]DownloadStationTask, error) {

	params := map[string]string{
		"api":        "SYNO.DownloadStation.Task",
		"version":    "1",
		"method":     "list",
		"additional": "transfer,detail",
	}

	resp, err := client.Get("webapi/DownloadStation/task.cgi", params)
	if err != nil {
		return nil, err
	}

	tasks := client.GetData(resp).(map[string]interface{})["tasks"].([]interface{})
	downloadTasks := mapDownloadStationTask(tasks)

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

func mapDownloadStationTask(tasks []interface{}) []DownloadStationTask {
	var downloadTasks []DownloadStationTask
	for _, task := range tasks {
		t := task.(map[string]interface{})
		transferInfo := t["additional"].(map[string]interface{})["transfer"].(map[string]interface{})
		detailInfo := t["additional"].(map[string]interface{})["detail"].(map[string]interface{})
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
					TaskDetail: TaskDetail{
						Destination: detailInfo["destination"].(string),
						Uri:         detailInfo["uri"].(string),
					},
				},
			})
	}
	return downloadTasks
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
