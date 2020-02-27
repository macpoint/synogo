package synoclient

import (
	"errors"
	"fmt"
	"sync"
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

type TaskAddError struct {
	Name string
	Err  error
}

var DsSynoErrors = map[int]string{
	400: "File upload failed",
	401: "Max number of tasks reached",
	402: "Destination denied",
	403: "Destination does not exist",
	404: "Invalid task id",
	405: "Invalid task action",
	406: "No default destination",
	407: "Set destination failed",
	408: "File does not exist",
}

// GetDownloadStationTask returns one DownloadStationTask
func (c *Client) GetDownloadStationTask(taskID string) (DownloadStationTask, error) {
	var dsTask DownloadStationTask
	tasksMap, err := c.GetDownloadStationTasks(taskID)
	if err != nil {
		return dsTask, err
	}
	for i := range tasksMap {
		if tasksMap[i].ID == taskID {
			dsTask = tasksMap[i]
		}
	}

	if dsTask.ID == "" {
		return dsTask, errors.New("Task not found")
	}

	return dsTask, nil
}

// GetDownloadStationTasks returns download tasks using 'getinfo'
func (c *Client) GetDownloadStationTasks(taskIds string) ([]DownloadStationTask, error) {
	params := map[string]string{
		"api":     "SYNO.DownloadStation.Task",
		"version": "1",
		"method":  "getinfo",
		// SynoAPI accepts multiple IPs separated by comma
		"id":         taskIds,
		"additional": "transfer,detail",
	}

	resp, err := c.Get("webapi/DownloadStation/task.cgi", params)
	if err != nil {
		return nil, HandleApplicationError(resp, err, DsSynoErrors)
	}

	tasks := c.GetData(resp).(map[string]interface{})["tasks"].([]interface{})
	downloadTasks := mapDownloadStationTask(tasks)

	return downloadTasks, nil

}

func (c *Client) ListDownloadStationTasks() ([]DownloadStationTask, error) {

	params := map[string]string{
		"api":        "SYNO.DownloadStation.Task",
		"version":    "1",
		"method":     "list",
		"additional": "transfer,detail",
	}

	resp, err := c.Get("webapi/DownloadStation/task.cgi", params)
	if err != nil {
		return nil, HandleApplicationError(resp, err, DsSynoErrors)
	}

	tasks := c.GetData(resp).(map[string]interface{})["tasks"].([]interface{})
	downloadTasks := mapDownloadStationTask(tasks)

	return downloadTasks, nil

}

func (c *Client) CreateDownloadStationTask(fileQueue <-chan string, errorQueue chan<- *TaskAddError, wg *sync.WaitGroup) error {

	params := map[string]string{
		"api":     "SYNO.DownloadStation.Task",
		"version": "1",
		"method":  "create",
	}
	defer wg.Done()
	for filename := range fileQueue {
		params["uri"] = filename
		fmt.Printf("Adding %v\n", truncateString(filename, 70))
		resp, err := c.Get("webapi/DownloadStation/task.cgi", params)
		if err != nil {
			errorQueue <- &TaskAddError{Name: filename, Err: HandleApplicationError(resp, err, DsSynoErrors)}
		}
	}
	return nil
}

func (c *Client) DeleteDownloadStationTasks(taskIds string) (response string, err error) {
	params := map[string]string{
		"api":     "SYNO.DownloadStation.Task",
		"version": "1",
		"method":  "delete",
		"id":      taskIds,
	}
	resp, err := c.Get("webapi/DownloadStation/task.cgi", params)
	if err != nil {
		return "", HandleApplicationError(resp, err, DsSynoErrors)
	}
	return resp, nil
}

func (c *Client) PauseDownloadStationTasks(taskIds string) (response string, err error) {
	params := map[string]string{
		"api":     "SYNO.DownloadStation.Task",
		"version": "1",
		"method":  "pause",
		"id":      taskIds,
	}
	resp, err := c.Get("webapi/DownloadStation/task.cgi", params)
	if err != nil {
		return "", HandleApplicationError(resp, err, DsSynoErrors)
	}
	return resp, nil
}

func (c *Client) ResumeDownloadStationTasks(taskIds string) (response string, err error) {
	params := map[string]string{
		"api":     "SYNO.DownloadStation.Task",
		"version": "1",
		"method":  "resume",
		"id":      taskIds,
	}
	resp, err := c.Get("webapi/DownloadStation/task.cgi", params)
	if err != nil {
		return "", HandleApplicationError(resp, err, DsSynoErrors)
	}
	return resp, nil
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
