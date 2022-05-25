package jobs

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cszatmary/fridge-monitor/monitorit/lib/sms"
	"github.com/cszatmary/fridge-monitor/monitorit/models"
)

type AlertJob struct {
	fm          *models.FridgeManager
	tm          *models.TemperatureManager
	smsClient   *sms.Client
	phoneNumber string
}

func NewAlertJob(fm *models.FridgeManager, tm *models.TemperatureManager, smsClient *sms.Client, phoneNumber string) *AlertJob {
	return &AlertJob{fm, tm, smsClient, phoneNumber}
}

func (aj *AlertJob) Run() {
	// Go through each fridge and perform the necessary checks
	ctx := context.Background()
	fridges, err := aj.fm.FindAll(ctx)
	if err != nil {
		// TODO(@cszatmary): Think about how to handle errors.
		// We need some way to surface this since if this fails then we won't get alerts.
		// Could potentially send a text on job failure but that might be too spammy.
		aj.alert("Failed to retrieve fridges: %v", err)
		return
	}
	for _, f := range fridges {
		if err := aj.checkFridge(ctx, f); err != nil {
			aj.alert("Failed to check fridge %s: %v", f.Name, err)
		}
	}
}

func (aj *AlertJob) checkFridge(ctx context.Context, fridge models.Fridge) error {
	// TODO(@cszatmary): We should probably make this configurable.
	const numTemps = 3
	// Get the last n temperatures which will be used to perform checks.
	temps, err := aj.tm.FindMostRecentByFridgeID(ctx, fridge.ID, numTemps)
	if err != nil {
		return err
	}

	// First check to make sure that a temperature was received in the expected interval.
	// TODO(@cszatmary): We should probably make the interval configurable.
	// Just use 30min for now since this seems reasonable.
	var lastReceived time.Time
	if len(temps) > 0 {
		lastReceived = temps[0].CreatedAt.Time
	}
	if time.Now().Sub(lastReceived) >= 30*time.Minute {
		// Have not received a temperature in the expected interval, alert!
		timeStr := "never"
		if !lastReceived.IsZero() {
			timeStr = lastReceived.Format(models.TimeFormatPretty)
		}
		aj.alert("Temperature not received from fridge %q since %s", fridge.Name, timeStr)
	}

	// Now check to make sure that the last n temperatures have been within the safe range.
	// All n temps must be outside to range to trigger an alert to avoid false alarms.

	// If we don't have n temperatures recorded yet then hold off, we need more data before we can be sure.
	if len(temps) < numTemps {
		log.Printf("AlertJob: Only have %d temperatures, waiting for more before checking status", len(temps))
		return nil
	}

	status := temps[0].Status(fridge.MinTemp, fridge.MaxTemp)
	if status == models.StatusNormal {
		// If latest is normal then all is good even if the previous ones aren't since it has either
		// recovered from a bad state or it was a flake.
		return nil
	}

	// We have a bad status, check the other two to see if they are bad as well.
	ok := false
	for _, t := range temps[1:] {
		if t.Status(fridge.MinTemp, fridge.MaxTemp) == models.StatusNormal {
			ok = true
			break
		}
	}

	if ok {
		// Not all n temperatures were bad so wait and see what happens.
		return nil
	}

	// All n temperatures are bad, we are in the danger zone, alert!
	var statusStr string
	var thresholdTemp string
	switch status {
	case models.StatusTooLow:
		statusStr = "too low"
		thresholdTemp = fmt.Sprintf("minimum safe temperature is %.2f°C", fridge.MinTemp)
	case models.StatusTooHigh:
		statusStr = "too high"
		thresholdTemp = fmt.Sprintf("maximum safe temperature is %.2f°C", fridge.MaxTemp)
	}
	aj.alert("Temperature of fridge %q is %s, current temperature is %.2f°C, %s", fridge.Name, statusStr, temps[0].Value, thresholdTemp)
	return nil
}

// alert performs an alert by both logging the message and sending an SMS.
func (aj *AlertJob) alert(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	log.Print("AlertJob: " + msg)
	if err := aj.smsClient.SendMessage(aj.phoneNumber, "MonitorIt: "+msg); err != nil {
		// Nothing we can realistically do here besides log it
		log.Printf("AlertJob Error: %v", err)
	}
}
