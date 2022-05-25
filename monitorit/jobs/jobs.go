package jobs

import (
	"time"

	"github.com/cszatmary/fridge-monitor/monitorit/lib/sms"
	"github.com/cszatmary/fridge-monitor/monitorit/models"
	"github.com/go-co-op/gocron"
)

type SetupDependencies struct {
	AlertJobCron        string
	FridgeManager       *models.FridgeManager
	TemperatureManager  *models.TemperatureManager
	SMSClient           *sms.Client
	AlertJobPhoneNumber string
}

func Setup(deps SetupDependencies) *gocron.Scheduler {
	s := gocron.NewScheduler(time.UTC)
	aj := NewAlertJob(deps.FridgeManager, deps.TemperatureManager, deps.SMSClient, deps.AlertJobPhoneNumber)
	s.Cron(deps.AlertJobCron).Do(aj.Run)
	return s
}
