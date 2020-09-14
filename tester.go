package tester

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
)

func Tester() {
	// <-- step1 : validate
	testCaseCount, err := strconv.ParseInt(os.Getenv("CASE_COUNT"), 10, 32)
	if err != nil {
		panic(err)
	}
	timeLimit, err := strconv.ParseInt(os.Getenv("TIME_LIMIT"), 10, 32)
	if err != nil {
		panic(err)
	}
	spaceLimit, err := strconv.ParseInt(os.Getenv("SPACE_LIMIT"), 10, 32)
	if err != nil {
		panic(err)
	}

	// todo: optimistic ? can we believe the scheduler and do less routine ???
	if testCaseCount <= 0 {
		panic(errors.New("invalid test case"))
	}

	for i := int64(1); i <= testCaseCount; i++ {
		if !Exists(DockerCasePath(i)) {
			panic(errors.New(fmt.Sprintf("Case #%d doesn't exist", i)))
		}
	}

	execCommandRaw := os.Getenv("EXEC_COMMAND")
	if execCommandRaw == "" {
		panic(err)
	}
	var execCommandArr []string
	if err := json.Unmarshal([]byte(execCommandRaw), &execCommandArr); err != nil {
		panic(err)
	}
	execCommand, execArgs := execCommandArr[0], execCommandArr[1:]

	if len(execCommandArr) == 1 {
		if err := os.Chmod(execCommandArr[0], 0755); err != nil {
			log.Println(err)
		}
	}

	file, err := os.Create(DockerResultFile())
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = file.Close()
	}()

	// <-- step2 : get_result
	testResult := make([]TestResult, testCaseCount)
	for i := int64(1); i <= testCaseCount; i++ {
		fmt.Printf("Test #%d Case...\n", i)
		TestOne(&testResult[i-1], i, timeLimit, spaceLimit, execCommand, execArgs)
	}

	// <-- step3 : write info
	result, err := json.Marshal(testResult)
	if err != nil {
		panic(err)
	}
	if _, err := file.Write(result); err != nil {
		panic(err)
	}

	os.Exit(0)
}
