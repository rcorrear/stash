package cli

import (
	"strconv"
	"time"

	sapi "github.com/appscode/stash/api"
	shell "github.com/codeskyblue/go-sh"
)

const (
	Exe = "/bin/restic"
)

type ResticWrapper struct {
	sh         *shell.Session
	scratchDir string
	hostname   string
}

func New(scratchDir string, hostname string) *ResticWrapper {
	ctrl := &ResticWrapper{
		sh:         shell.NewSession(),
		scratchDir: scratchDir,
		hostname:   hostname,
	}
	ctrl.sh.SetDir(scratchDir)
	ctrl.sh.ShowCMD = true
	return ctrl
}

type Snapshot struct {
	ID       string    `json:"id"`
	Time     time.Time `json:"time"`
	Tree     string    `json:"tree"`
	Paths    []string  `json:"paths"`
	Hostname string    `json:"hostname"`
	Username string    `json:"username"`
	UID      int       `json:"uid"`
	Gid      int       `json:"gid"`
}

func (w *ResticWrapper) ListSnapshots() ([]Snapshot, error) {
	result := make([]Snapshot, 0)
	err := w.sh.Command(Exe, "snapshots", "--json").UnmarshalJSON(&result)
	return result, err
}

func (w *ResticWrapper) InitRepositoryIfAbsent() error {
	if err := w.sh.Command(Exe, "snapshots", "--json").Run(); err != nil {
		return w.sh.Command(Exe, "init").Run()
	}
	return nil
}

func (w *ResticWrapper) Backup(resource *sapi.Restic, fg sapi.FileGroup) error {
	args := []interface{}{"backup", fg.Path, "--force"}
	// add tags if any
	for _, tag := range fg.Tags {
		args = append(args, "--tag")
		args = append(args, tag)
	}
	return w.sh.Command(Exe, args...).Run()
}

func (w *ResticWrapper) Forget(resource *sapi.Restic, fg sapi.FileGroup) error {
	args := []interface{}{"forget"}
	if fg.RetentionPolicy.KeepLastSnapshots > 0 {
		args = append(args, string(sapi.KeepLast))
		args = append(args, strconv.Itoa(fg.RetentionPolicy.KeepLastSnapshots))
	}
	if fg.RetentionPolicy.KeepHourlySnapshots > 0 {
		args = append(args, string(sapi.KeepHourly))
		args = append(args, strconv.Itoa(fg.RetentionPolicy.KeepHourlySnapshots))
	}
	if fg.RetentionPolicy.KeepDailySnapshots > 0 {
		args = append(args, string(sapi.KeepDaily))
		args = append(args, strconv.Itoa(fg.RetentionPolicy.KeepDailySnapshots))
	}
	if fg.RetentionPolicy.KeepWeeklySnapshots > 0 {
		args = append(args, string(sapi.KeepWeekly))
		args = append(args, strconv.Itoa(fg.RetentionPolicy.KeepWeeklySnapshots))
	}
	if fg.RetentionPolicy.KeepMonthlySnapshots > 0 {
		args = append(args, string(sapi.KeepMonthly))
		args = append(args, strconv.Itoa(fg.RetentionPolicy.KeepMonthlySnapshots))
	}
	if fg.RetentionPolicy.KeepYearlySnapshots > 0 {
		args = append(args, string(sapi.KeepYearly))
		args = append(args, strconv.Itoa(fg.RetentionPolicy.KeepYearlySnapshots))
	}
	for _, tag := range fg.RetentionPolicy.KeepTags {
		args = append(args, string(sapi.KeepTag))
		args = append(args, tag)
	}
	for _, tag := range fg.Tags {
		args = append(args, "--tag")
		args = append(args, tag)
	}
	err := w.sh.Command(Exe, args...).Run()
	if err != nil {
		return err
	}
	return nil
}