package tester

import (
	"fmt"
	"os"
)

func DockerCasePath(caseId int64) string {
	return fmt.Sprintf("/case/%d.in", caseId)
}

func DockerOutputPath(caseId int64) string {
	return fmt.Sprintf("/output/%d.out", caseId)
}

func DockerResultFile() string {
	return "/result/info.json"
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}