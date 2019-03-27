package backup

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/codeskyblue/go-sh"
	"github.com/pkg/errors"
	"github.com/stefanprodan/mgob/config"
)

func dump(plan config.Plan, tmpPath string, ts time.Time) (string, string, error) {
	switch plan.Target.Platform {
	case "influxdb":
		return dumpInfluxDB(plan, tmpPath, ts)
	case "prometheus":
		return dumpPrometheus(plan, tmpPath, ts)
	case "gitlab":
		return dumpGitlab(plan, tmpPath, ts)
	case "mongodb":
		target, err := config.LoadMongoDBTarget("/secrets", plan.Target.ExistingSecret)
		if err != nil {
			return "", "", errors.Wrapf(err, "Cannot load target %v", plan.Target.ExistingSecret)
		}
		return dumpMongo(plan, target, tmpPath, ts)
	}
	return "", fmt.Sprintf("Unknown platform %v", plan.Target.Platform), nil

}

func dumpInfluxDB(plan config.Plan, tmpPath string, ts time.Time) (string, string, error) {
		return "", "", errors.Wrapf(nil, "Not implemented", "")
}

func dumpPrometheus(plan config.Plan, tmpPath string, ts time.Time) (string, string, error) {
		return "", "", errors.Wrapf(nil, "Not implemented", "")
}

func dumpGitlab(plan config.Plan, tmpPath string, ts time.Time) (string, string, error) {
		return "", "", errors.Wrapf(nil, "Not implemented", "")
}

func dumpMongo(plan config.Plan, target config.MongoDBTarget, tmpPath string, ts time.Time) (string, string, error) {
	archive := fmt.Sprintf("%v/%v-%v.gz", tmpPath, plan.Name, ts.Unix())
	log := fmt.Sprintf("%v/%v-%v.log", tmpPath, plan.Name, ts.Unix())

	dump := fmt.Sprintf("mongodump --archive=%v --gzip --host %v --port %v ",
		archive, target.Host, target.Port)
	if target.Database != "" {
		dump += fmt.Sprintf("--db %v ", target.Database)
	}
	if target.Username != "" && target.Password != "" {
		dump += fmt.Sprintf("-u %v -p %v ", target.Username, target.Password)
	}
	if target.Params != "" {
		dump += fmt.Sprintf("%v", target.Params)
	}

	output, err := sh.Command("/bin/sh", "-c", dump).SetTimeout(time.Duration(plan.Scheduler.Timeout) * time.Minute).CombinedOutput()
	if err != nil {
		ex := ""
		if len(output) > 0 {
			ex = strings.Replace(string(output), "\n", " ", -1)
		}
		return "", "", errors.Wrapf(err, "mongodump log %v", ex)
	}
	logToFile(log, output)

	return archive, log, nil
}

func logToFile(file string, data []byte) error {
	if len(data) > 0 {
		err := ioutil.WriteFile(file, data, 0644)
		if err != nil {
			return errors.Wrapf(err, "writing log %v failed", file)
		}
	}

	return nil
}

func applyRetention(path string, retention int) error {
	gz := fmt.Sprintf("cd %v && rm -f $(ls -1t *.gz | tail -n +%v)", path, retention+1)
	err := sh.Command("/bin/sh", "-c", gz).Run()
	if err != nil {
		return errors.Wrapf(err, "removing old gz files from %v failed", path)
	}

	log := fmt.Sprintf("cd %v && rm -f $(ls -1t *.log | tail -n +%v)", path, retention+1)
	err = sh.Command("/bin/sh", "-c", log).Run()
	if err != nil {
		return errors.Wrapf(nil, "removing old log files from %v failed", path)
	}

	return nil
}

// TmpCleanup remove files older than one day
func TmpCleanup(path string) error {
	rm := fmt.Sprintf("find %v -not -name \"mgob.db\" -mtime +%v -type f -delete", path, 1)
	err := sh.Command("/bin/sh", "-c", rm).Run()
	if err != nil {
		return errors.Wrapf(err, "%v cleanup failed", path)
	}

	return nil
}
