package alert

import (
	"github.com/boivie/lovebeat/config"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type scriptAlerter struct {
}

func (m scriptAlerter) Notify(cfg config.ConfigAlert, ev AlertInfo) {
	if cfg.Script != "" {
		log.Info("Running alert script %s", cfg.Script)
		cmd := exec.Command(cfg.Script)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = append(os.Environ(), []string{
			"LOVEBEAT_ALARM=" + ev.Alarm.Name,
			"LOVEBEAT_STATE=" + strings.ToUpper(ev.Current),
			"LOVEBEAT_PREVIOUS_STATE=" + strings.ToUpper(ev.Previous),
			"LOVEBEAT_INCIDENT=" + strconv.Itoa(ev.Alarm.IncidentNbr)}...)
		c := make(chan int, 1)
		go func() {
			err := cmd.Run()
			if err != nil {
				log.Warning("Finished with error code: %v", err)
			}
			c <- 1
		}()
		select {
		case <-c:
		// ok
		case <-time.After(10 * time.Second):
			log.Warning("Timed out waiting for script to exit")
			cmd.Process.Kill()
		}
	}
}

func NewScriptAlerter(cfg config.Config) AlerterBackend {
	return &scriptAlerter{}
}
