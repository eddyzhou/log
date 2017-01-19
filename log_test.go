package log

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestLog(t *testing.T) {
	Debug("Debug: foo")
	Debugf("Debug: %s", "foo")
	Info("Info: foo")
	Infof("Info: %s", "foo")
	Warn("Warn: foo")
	Warnf("Warn: %s", "foo")
	Error("Error: foo")
	Errorf("Error: %s", "foo")

	Std = New(os.Stderr, "", Ldefault, Linfo)
	Debug("Debug: bar")
	Debugf("Debug: %s", "bar")
	Info("Info: bar")
	Infof("Info: %s", "bar")
	Warn("Warn: bar")
	Warnf("Warn: %s", "bar")
	Error("Error: bar")
	Errorf("Error: %s", "bar")
}

func TestRotate(t *testing.T) {
	Info("Test Rotate...")

	fileName := "test.log"
	if isFileExist(fileName) {
		if err := os.Remove(fileName); err != nil {
			t.Errorf("remove old log file failed: %s", err.Error())
		}
	}

	Std = NewRotate(fileName, "", Ldefault, Linfo)

	now := time.Now()
	suffix := now.Format(TimeLayout)
	t.Log("suffix:", suffix)
	if Std.timeSuffix != suffix {
		t.Errorf("bad timeSuffix: %s", suffix)
	}

	Info("log 1")
	Info("log 2")

	if err := checkLogData(fileName, "log 1", "log 2", 2); err != nil {
		t.Errorf("check log data failed: %s", err.Error())
	}

	lastDay := now.AddDate(0, 0, -1)
	lastSuffix := lastDay.Format(TimeLayout)
	Std.timeSuffix = lastSuffix

	Info("log 3")
	Info("log 4")

	if err := Std.fd.Close(); err != nil {
		t.Errorf("close log fd[%s] failed: %s", Std.fileName, err.Error())
	}

	oldLogName := fileName + "." + lastSuffix
	if err := checkLogData(oldLogName, "log 1", "log 2", 2); err != nil {
		t.Errorf("check old log data failed: %s", err.Error())
	}

	if err := checkLogData(fileName, "log 3", "log 4", 2); err != nil {
		t.Errorf("check log data failed: %s", err.Error())
	}

	if err := os.Remove(oldLogName); err != nil {
		t.Errorf("remove old log file failed: %s", err.Error())
	}

	if err := os.Remove(fileName); err != nil {
		t.Errorf("remove log file failed: %s", err.Error())
	}

}

func checkLogData(fileName, firstLine, lastLine string, totalLine int) error {
	input, err := os.OpenFile(fileName, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer input.Close()

	var last string
	var num int
	var firstChecked bool
	br := bufio.NewReader(input)
	for {
		line, err := br.ReadString('\n')
		if err == io.EOF {
			break
		}

		line = strings.TrimRight(line, "\n")
		last = line
		num++

		if !firstChecked {
			if strings.Contains(firstLine, line) {
				return fmt.Errorf("first line except %s but %s", firstLine, line)
			}
			firstChecked = true
		}
	}

	if strings.Contains(lastLine, last) {
		return fmt.Errorf("last line except %s but %s", lastLine, last)
	}

	if totalLine != num {
		return fmt.Errorf("totalLine except %d but %d", totalLine, num)
	}

	return nil
}

func isFileExist(name string) bool {
	f, err := os.Stat(name)
	if err != nil && os.IsNotExist(err) {
		return false
	}

	if f.IsDir() {
		return false
	}

	return true
}
